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
	"github.com/cockroachdb/cockroach-operator/pkg/update"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *clusterRestart) Handles(conds []api.ClusterCondition) bool {
	return utilfeature.DefaultMutableFeatureGate.Enabled(features.ClusterRestart) && condition.True(api.InitializedCondition, conds) || condition.False(api.InitializedCondition, conds)
}

func (r *clusterRestart) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := r.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.V(int(zapcore.DebugLevel)).Info("starting cluster restart action")
	if cluster.Spec().RestartType == "" {
		log.V(int(zapcore.DebugLevel)).Info("No restart cluster action")
		return nil
	}
	// Get the sts and compare the sts size to the size in the CR
	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      cluster.StatefulSetName(),
	}
	statefulSet := &appsv1.StatefulSet{}
	if err := r.client.Get(ctx, key, statefulSet); err != nil {
		return errors.Wrap(err, "failed to fetch statefulset")
	}
	clientset, err := kubernetes.NewForConfig(r.config)
	if err != nil {
		return errors.Wrapf(err, "failed to create kubernetes clientset")
	}

	if cluster.Spec().RestartType == api.RollingRestart {
		log.V(int(zapcore.DebugLevel)).Info("initiating rolling restart action")
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
		timeNow := metav1.Now()
		// stsName := statefulSet.Name
		// stsNamespace := statefulSet.Namespace
		//this annotation will trigger the rolling update
		statefulSet.Annotations[resource.CrdbRestartAnnotation] = timeNow.Format(time.RFC3339)
		// _, err := clientset.AppsV1().StatefulSets(stsNamespace).Update(ctx, statefulSet, metav1.UpdateOptions{})

		// if err != nil {
		// 	return handleStsError(err, log, stsName, stsNamespace)
		// }
		log.V(int(zapcore.DebugLevel)).Info("BEFORE rolling")
		if err := r.RollingSts(ctx, cluster, clientset); err != nil {
			return errors.Wrapf(err, "error restarting statefulset %s.%s", cluster.Namespace(), cluster.StatefulSetName())
		}
		log.V(int(zapcore.DebugLevel)).Info("completed rolling cluster restart")
		return nil
	}

	if cluster.Spec().RestartType == api.FullRestart {
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

		if err := r.ScaleSts(ctx, statefulSet, log, clientset, int32(0)); err != nil {
			return errors.Wrapf(err, "error reseting statefulset %s.%s to 0 replicas", cluster.Namespace(), cluster.StatefulSetName())
		}
		log.V(int(zapcore.DebugLevel)).Info("completed setting the statefullset replicas to 0")

		replicas := cluster.Spec().Nodes

		if err := r.ScaleSts(ctx, statefulSet, log, clientset, replicas); err != nil {
			return errors.Wrapf(err, "error reseting statefulset %s.%s to %v replicas", cluster.Namespace(), cluster.StatefulSetName(), replicas)
		}

		log.V(int(zapcore.DebugLevel)).Info("completed full cluster restart")
		return nil
	}

	log.V(int(zapcore.DebugLevel)).Info("completed cluster restart")
	return nil
}

// RollingSts performs a rolling update on the cluster.
func (r *clusterRestart) RollingSts(ctx context.Context, cluster *resource.Cluster, clientset *kubernetes.Clientset) error {
	updateRoach := &update.UpdateRoach{
		StsName:      cluster.StatefulSetName(),
		StsNamespace: cluster.Namespace(),
	}

	podUpdateTimeout := 10 * time.Minute
	podMaxPollingInterval := 30 * time.Minute
	sleeper := update.NewSleeper(1 * time.Minute)

	k8sCluster := &update.UpdateCluster{
		Clientset:             clientset,
		PodUpdateTimeout:      podUpdateTimeout,
		PodMaxPollingInterval: podMaxPollingInterval,
		Sleeper:               sleeper,
	}
	return update.RollingRestart(ctx, updateRoach, k8sCluster, r.log)
}

//ScaleSts updates the replicas of the statefullset to the value from parameters
//and waits until the statefullset has currentreplicas equal with the desired replicas
//We will use this in FullRestart logic, by setting replicas to 0 and
//afterwords setting to the original number of nodes, this way the cluster
//will load the new CA certs
func (r *clusterRestart) ScaleSts(ctx context.Context, sts *appsv1.StatefulSet, l logr.Logger, clientset *kubernetes.Clientset, replicas int32) error {
	timeNow := metav1.Now()
	stsName := sts.Name
	stsNamespace := sts.Namespace
	sts.Spec.Replicas = &replicas
	sts.Annotations[resource.CrdbRestartAnnotation] = timeNow.Format(time.RFC3339)
	_, err := clientset.AppsV1().StatefulSets(stsNamespace).Update(ctx, sts, metav1.UpdateOptions{})

	if err != nil {
		return handleStsError(err, l, stsName, stsNamespace)
	}
	return scale.WaitUntilStatefulSetIsReadyToServe(ctx, clientset, stsNamespace, stsName, int32(replicas))
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
