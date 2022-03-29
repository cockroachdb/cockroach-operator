/*
Copyright 2022 The Cockroach Authors

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
	"github.com/go-logr/logr"
	"k8s.io/client-go/util/retry"
	"strconv"
	"strings"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newInitialize(scheme *runtime.Scheme, cl client.Client, config *rest.Config, clientset kubernetes.Interface) Actor {
	return &initialize{
		action: newAction(scheme, cl, config, clientset),
	}
}

// initialize performs the initialization of the new cluster
type initialize struct {
	action
}

// GetActionType returns the  api.InitializeAction value used to set the cluster status errors
func (init initialize) GetActionType() api.ActionType {
	return api.InitializeAction
}

func (init initialize) Act(ctx context.Context, cluster *resource.Cluster, log logr.Logger) error {
	log.V(DEBUGLEVEL).Info("initializing CockroachDB")

	stsName := cluster.StatefulSetName()

	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      stsName,
	}
	ss := &appsv1.StatefulSet{}
	if err := init.client.Get(ctx, key, ss); err != nil {
		log.Error(err, "failed to fetch statefulset")
		return kube.IgnoreNotFound(err)
	}

	pods, err := init.clientset.CoreV1().Pods(ss.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(ss.Spec.Selector.MatchLabels).AsSelector().String(),
	})

	if err != nil {
		msg := "error getting pods in statefulset"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	}

	if len(pods.Items) == 0 {
		return NotReadyErr{Err: errors.New("pod not created")}
	}

	phase := pods.Items[0].Status.Phase
	podName := pods.Items[0].Name
	if phase != corev1.PodRunning {
		return NotReadyErr{Err: errors.New("pod is not running")}
	}

	log.V(DEBUGLEVEL).Info("Pod is ready")

	port := strconv.FormatInt(int64(*cluster.Spec().GRPCPort), 10)
	cmd := []string{
		"/cockroach/cockroach.sh",
		"init",
		cluster.SecureMode(),
		"--host=localhost:" + port,
	}

	log.V(DEBUGLEVEL).Info(fmt.Sprintf("Executing init in pod %s with phase %s", podName, phase))
	_, stderr, err := kube.ExecInPod(init.scheme, init.config, cluster.Namespace(),
		fmt.Sprintf("%s-0", stsName), resource.DbContainerName, cmd)
	log.V(DEBUGLEVEL).Info("Executed init in pod")

	if err != nil && !alreadyInitialized(stderr) {
		// can happen if container has not finished its startup
		if strings.Contains(err.Error(), "unable to upgrade connection: container not found") ||
			strings.Contains(err.Error(), "does not have a host assigned") {
			log.V(DEBUGLEVEL).Info("pod has not completely started")
			return NotReadyErr{Err: errors.New("pod has not completely started")}
		}

		msg := "failed to initialize the cluster"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	}

	// If we got here, we need to update the CrdbClusterStatus object with an updated CrdbInitialized condition.
	fetcher := resource.NewKubeFetcher(ctx, cluster.Namespace(), init.client)
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch the latest CrdbClusterStatus
		newcr := resource.ClusterPlaceholder(cluster.Name())
		if err := fetcher.Fetch(newcr); err != nil {
			msg := "failed to retrieve CrdbCluster resource"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}

		// Set the CrdbInitialized condition in the cluster status to true
		refreshedCluster := resource.NewCluster(newcr)
		refreshedCluster.Fetcher = fetcher
		refreshedCluster.SetTrue(api.CrdbInitializedCondition)

		// Actually attempt to update the CrdbClusterStatus object. If the update runs into a conflict for any reason
		// (say, someone adds a label to the CrdbCluster object after we just retrieved it), we will retry.
		err = init.client.Status().Update(ctx, refreshedCluster.Unwrap())
		if err != nil {
			msg := "failed to update initialized annotation; will try again"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}
		return err
	})
	if err != nil {
		msg := "failed to update initialized annotation"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	}

	log.V(DEBUGLEVEL).Info("completed initializing database")
	return nil
}

func alreadyInitialized(out string) bool {
	return strings.Contains(out, "cluster has already been initialized")
}
