/*
Copyright 2021 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package actor

import (
	"context"
	"fmt"
	"time"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/scale"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newClusterRestart(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &clusterRestart{
		action: newAction("Crdb Cluster Restart", scheme, cl),
		config: config,
	}
}

// clusterRestart will restart the CRDB cluster using 2 option: Rolling Restart and
// Full Restart in case of CA renew
type clusterRestart struct {
	action

	config *rest.Config
}

//GetActionType returns api.ClusterRestartAction action used to set the cluster status errors
func (r *clusterRestart) GetActionType() api.ActionType {
	return api.ClusterRestartAction
}

//Handles will return true if the prerequisite are met to run restart
//like the cluster exists, actin deploy and inizialize, run and the feature gate was enabled
//for this feature
func (r *clusterRestart) Handles(conds []api.ClusterCondition) bool {
	return utilfeature.DefaultMutableFeatureGate.Enabled(features.ClusterRestart) &&
		(condition.True(api.InitializedCondition, conds) || condition.False(api.InitializedCondition, conds)) &&
		condition.True(api.CrdbVersionChecked, conds)
}

func (r *clusterRestart) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := r.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.V(int(zapcore.DebugLevel)).Info("starting cluster restart action")
	restartType := cluster.GetAnnotationRestartType()
	if restartType == "" {
		log.V(int(zapcore.DebugLevel)).Info("No restart cluster action")
		return nil
	}
	// Get the sts and compare the sts size to the size in the CR
	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      cluster.StatefulSetName(),
	}
	clientset, err := kubernetes.NewForConfig(r.config)
	if err != nil {
		return errors.Wrapf(err, "failed to create kubernetes clientset")
	}
	statefulSet := &appsv1.StatefulSet{}
	if err := r.client.Get(ctx, key, statefulSet); err != nil {
		return errors.Wrap(err, "failed to fetch statefulset")
	}
	// TODO statefulSetIsUpdating is not quite working as expected.
	// I had to check status.  We should look at the update code in partition update to address this
	if statefulSetIsUpdating(statefulSet) {
		return NotReadyErr{Err: errors.New("restart statefulset is updating, waiting for the update to finish")}
	}

	status := &statefulSet.Status
	if status.CurrentReplicas == 0 || status.CurrentReplicas < status.Replicas {
		log.Info("restart statefulset does not have all replicas up")
		return NotReadyErr{Err: errors.New("restart cluster statefulset does not have all replicas up")}
	}

	if restartType == api.ClusterRestartType(api.RollingRestart).String() {
		log.V(int(zapcore.DebugLevel)).Info("initiating rolling restart action")
		if err := r.rollingSts(ctx, statefulSet.DeepCopy(), clientset, r.log); err != nil {
			return errors.Wrapf(err, "error restarting statefulset %s.%s", cluster.Namespace(), cluster.StatefulSetName())
		}
		log.V(int(zapcore.DebugLevel)).Info("completed rolling cluster restart")
	} else if restartType == api.ClusterRestartType(api.FullCluster).String() {
		if err := r.fullClusterRestart(ctx, statefulSet, log, clientset); err != nil {
			return errors.Wrapf(err, "error reseting statefulset %s.%s to 0 replicas", cluster.Namespace(), cluster.StatefulSetName())
		}
		log.V(int(zapcore.DebugLevel)).Info("completed full cluster restart")
	} else {
		log.V(int(zapcore.DebugLevel)).Info("invalid annotation")
	}
	// we force the saving of the status on the cluster and cancel the loop
	fetcher := resource.NewKubeFetcher(ctx, cluster.Namespace(), r.client)

	cr := resource.ClusterPlaceholder(cluster.Name())
	if err := fetcher.Fetch(cr); err != nil {
		log.Error(err, "failed to retrieve CrdbCluster resource")
		return err
	}
	refreshedCluster := resource.NewCluster(cr)
	// save the status of the cluster, we mark as restarted the cluster
	refreshedCluster.SetActionFinished(api.ClusterRestartAction)
	//reset this... for now
	refreshedCluster.DeleteRestartTypeAnnotation()
	//we should patch
	if err := r.client.Update(ctx, refreshedCluster.Unwrap()); err != nil {
		log.Error(err, "failed resetting the restart cluster field")
	}
	if err := r.client.Status().Update(ctx, refreshedCluster.Unwrap()); err != nil {
		log.Error(err, "failed saving cluster status on cluster restart")
		return nil
	}

	log.V(int(zapcore.DebugLevel)).Info("completed cluster restart")
	return nil
}

// rollingSts performs a rolling update on the cluster.
func (r *clusterRestart) rollingSts(ctx context.Context, sts *appsv1.StatefulSet, clientset *kubernetes.Clientset, l logr.Logger) error {
	timeNow := metav1.Now()
	// When a StatefulSet's partition number is set to `n`, only StatefulSet pods
	// numbered greater or equal to `n` will be updated. The rest will remain untouched.
	// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#partitions
	for partition := *sts.Spec.Replicas - 1; partition >= 0; partition-- {
		stsName := sts.Name
		stsNamespace := sts.Namespace
		replicas := sts.Spec.Replicas

		refreshedSts, err := clientset.AppsV1().StatefulSets(stsNamespace).Get(ctx, stsName, metav1.GetOptions{})
		if err != nil {
			return handleStsError(err, l, stsName, stsNamespace)
		}
		sts := refreshedSts.DeepCopy()
		sts.Annotations[resource.CrdbRestartAnnotation] = timeNow.Format(time.RFC3339)

		if sts.Spec.Template.Annotations == nil {
			sts.Spec.Template.Annotations = make(map[string]string)
		}
		sts.Spec.Template.Annotations[resource.CrdbRestartAnnotation] = timeNow.Format(time.RFC3339)
		sts.Spec.UpdateStrategy.RollingUpdate = &v1.RollingUpdateStatefulSetStrategy{
			Partition: &partition,
		}
		_, err = clientset.AppsV1().StatefulSets(stsNamespace).Update(ctx, sts, metav1.UpdateOptions{})
		if err != nil {
			return handleStsError(err, l, stsName, stsNamespace)
		}
		// Wait until verificationFunction verifies the update, passing in
		// the current partition so the function knows which pod to check
		// the status of.
		l.V(int(zapcore.DebugLevel)).Info("waiting until partition done restarting", "partition number:", partition)

		if err := scale.WaitUntilStatefulSetIsReadyToServe(ctx, clientset, stsNamespace, stsName, *replicas); err != nil {
			return errors.Wrapf(err, "error rolling update stategy on pod %d", int(partition))
		}

		// wait 1 minute between updates
		duration := 1 * time.Minute
		l.V(int(zapcore.DebugLevel)).Info("sleeping", "duration", duration.String(), "label", "between restarting pods")
		time.Sleep(1 * time.Minute)
	}
	return nil
}

//fullClusterRestart will delete all the pods of the sts
//to force the reload of the certificateon the POD
//used on the CA cert rotation
func (r *clusterRestart) fullClusterRestart(ctx context.Context, sts *appsv1.StatefulSet, l logr.Logger, clientset *kubernetes.Clientset) error {

	timeNow := metav1.Now()
	stsName := sts.Name
	stsNamespace := sts.Namespace
	sts.Annotations[resource.CrdbRestartAnnotation] = timeNow.Format(time.RFC3339)

	_, err := clientset.AppsV1().StatefulSets(stsNamespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return handleStsError(err, l, stsName, stsNamespace)
	}
	dp := metav1.DeletePropagationForeground
	err = clientset.CoreV1().Pods(sts.Namespace).DeleteCollection(ctx, metav1.DeleteOptions{
		PropagationPolicy: &dp,
	}, metav1.ListOptions{
		LabelSelector: labels.Set(sts.Spec.Selector.MatchLabels).AsSelector().String(),
	})
	if err != nil {
		l.Error(err, "failed to delete the pods for sts")
		return err
	}
	//waiting for autohealing
	return scale.WaitUntilStatefulSetIsReadyToServe(ctx, clientset, stsNamespace, stsName, *sts.Spec.Replicas)
}

func handleStsError(err error, l logr.Logger, stsName string, ns string) error {
	if k8sErrors.IsNotFound(err) {
		l.Error(err, "sts is not found", "stsName", stsName, "namespace", ns)
		return errors.Wrapf(err, "sts is not found: %s ns: %s", stsName, ns)
	} else if statusError, isStatus := err.(*k8sErrors.StatusError); isStatus {
		l.Error(statusError, fmt.Sprintf("Error getting statefulset %v", statusError.ErrStatus.Message), "stsName", stsName, "namespace", ns)
		return statusError
	}
	l.Error(err, "error getting statefulset", "stsName", stsName, "namspace", ns)
	return err
}
