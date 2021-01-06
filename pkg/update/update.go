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

package update

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// The default maximum amount of time to wait between when a pod has been
	// updated and when it is safe to proceed to the next pod.
	defaultPodUpdateTimeout = 5 * time.Minute

	// The default maximum amount of time to wait between polling attempts when
	// checking if we're okay to proceed with the update.
	defaultPodMaxPollingInterval = 5 * time.Second

	PreserveDowngradeOptionClusterSetting = "cluster.preserve_downgrade_option"
)

// updateFunctionSuite is a collection of functions used to update the
// CockroachDB StatefulSet in each region of a CockroachDB cluster. This suite
// gets passed as an argument to updateClusterStatefulSets to handle the update
// on a specific cluster.
//
// updateFunc takes a StatefulSet model and applies the expected
// changes to the model. For example, it may change the `image` value of one of
// the containers within the CockroachDB StatefulSet. updateFunc should return
// the updated StatefulSet model.
//
// updateStrategyFunc takes a Kubernetes client, a region model,
// and the StatefulSet model which has been modified by updateFunc, and is
// responsible for rolling out the changes to the pods within that StatefulSet.
// If you don't know what function to use to apply your update, by default you
// should use partitionedRollingUpdateStrategy (defined in this package).
type updateFunctionSuite struct {
	updateFunc         func(sts *v1.StatefulSet) (*v1.StatefulSet, error)
	updateStrategyFunc func(update *UpdateSts, updateTimer *UpdateTimer, l logr.Logger) (bool, error)
}

// TODO consolidate structs. We have structs in update_version that mirror these

// UpdateSts struct encapsultates everything Kubernetes related we need in order to update
// a StatefulSet
type UpdateSts struct {
	ctx       context.Context
	clientset kubernetes.Interface
	sts       *v1.StatefulSet
	namespace string
	name      string
}

// UpdateTimer encapsulates everything timer and polling related we need to update
// a StatefulSet.
type UpdateTimer struct {
	podUpdateTimeout      time.Duration
	podMaxPollingInterval time.Duration
	sleeper               Sleeper
	// TODO check that this func is actually correct
	waitUntilAllPodsReadyFunc func(context.Context, logr.Logger) error
}

func NewUpdateFunctionSuite(
	updateFunc func(*v1.StatefulSet) (*v1.StatefulSet, error),
	updateStrategyFunc func(update *UpdateSts, updateTimer *UpdateTimer, l logr.Logger) (bool, error),
) *updateFunctionSuite {
	return &updateFunctionSuite{
		updateFunc:         updateFunc,
		updateStrategyFunc: updateStrategyFunc,
	}
}

type Sleeper interface { // for testing
	Sleep(l logr.Logger, logSuffix string)
}

type sleeperImpl struct {
	duration time.Duration
}

func NewSleeper(duration time.Duration) *sleeperImpl {
	return &sleeperImpl{duration: duration}
}

func (s *sleeperImpl) Sleep(l logr.Logger, logSuffix string) {
	l.Info("sleeping", "duration", string(s.duration), "label", logSuffix)
	time.Sleep(s.duration)
}

// TODO rewrite docs
// TODO too many parmeters, just found a bug where I reversed namespace and sts name
// Refactor this into a struct

// UpdateClusterRegionStatefulSet is the regional version of
// updateClusterStatefulSets. See its documentation for more information on the
// parameters passed to this function.
func UpdateClusterRegionStatefulSet(
	ctx context.Context,
	clientset kubernetes.Interface,
	name string,
	namespace string,
	updateSuite *updateFunctionSuite,
	waitUntilAllPodsReadyFunc func(context.Context, logr.Logger) error,
	podUpdateTimeout time.Duration,
	podMaxPollingInterval time.Duration,
	sleeper Sleeper,
	l logr.Logger,
) (bool, error) {
	l = l.WithName(namespace)

	sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, handleStsError(err, l, name, namespace)
	}

	// Run the updateFunc to update the in-memory copy of the Kubernetes
	// resource.  The new in-memory copy of the Kubernetes resource is not
	// applied to the cluster by updateFunc, that is handled by the
	// updateStrategyFunc.
	sts, err = updateSuite.updateFunc(sts)
	if err != nil {
		return false, errors.Wrapf(err, "error applying updateFunc to %s %s", name, namespace)
	}
	updateSts := &UpdateSts{
		ctx:       ctx,
		clientset: clientset,
		sts:       sts,
		name:      name,
		namespace: namespace,
	}

	updateTimer := &UpdateTimer{
		podUpdateTimeout:          podUpdateTimeout,
		podMaxPollingInterval:     podMaxPollingInterval,
		sleeper:                   sleeper,
		waitUntilAllPodsReadyFunc: waitUntilAllPodsReadyFunc,
	}
	// updateStrategyFunc is responsible for controlling the rollout of the
	// changed StatefulSet definition across the pods in the Statefulset.
	skipSleep, err := updateSuite.updateStrategyFunc(updateSts, updateTimer, l)
	if err != nil {
		return false, errors.Wrapf(err, "error applying updateStrategyFunc to %s %s", name, namespace)
	}

	return skipSleep, nil
}

// partitionedRollingUpdateStrategy is an update strategy which updates the
// pods in a statefulset one at a time, and verifies the health of the
// cluster throughout the update.
//
// partitionedRollingUpdateStrategy checks that all pods are ready before
// replacing a pod within a cluster.
//
// After a pod has been updated, the perPodVerificationFunc will run to ensure
// the pod is in the expected state before continuing the update. This function
// takes a Kubernetes clientset, the StatefulSet being modified, and the pod
// number of the Statefulset that has just been updated. If it returns an error,
// the update is halted.
func PartitionedRollingUpdateStrategy(perPodVerificationFunc func(*UpdateSts, int, logr.Logger) error,
) func(updateSts *UpdateSts, updateTimer *UpdateTimer, l logr.Logger) (bool, error) {
	return func(updateSts *UpdateSts, updateTimer *UpdateTimer, l logr.Logger) (bool, error) {
		// When a StatefulSet's partition number is set to `n`, only StatefulSet pods
		// numbered greater or equal to `n` will be updated. The rest will remain untouched.
		// https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#partitions
		skipSleep := false
		sts := updateSts.sts
		for partition := *sts.Spec.Replicas - 1; partition >= 0; partition-- {
			stsName := sts.Name
			stsNamespace := sts.Namespace

			// If pod already updated, we are probably retrying a failed job
			// attempt. Best not to redo the update in that case, especially the sleeps!!
			if err := perPodVerificationFunc(updateSts, int(partition), l); err == nil {
				l.Info("already updated, skipping sleep", "partition", partition)
				skipSleep = true
				continue
			}

			skipSleep = false
			// TODO we are only using this func here.  Why are we passing it around?
			if err := updateTimer.waitUntilAllPodsReadyFunc(updateSts.ctx, l); err != nil {
				return false, errors.Wrapf(err, "error while waiting for all pods to be ready")
			}
			sts.Spec.UpdateStrategy.RollingUpdate = &v1.RollingUpdateStatefulSetStrategy{
				Partition: &partition,
			}

			_, err := updateSts.clientset.AppsV1().StatefulSets(stsNamespace).Update(updateSts.ctx, sts, metav1.UpdateOptions{})
			if err != nil {
				return false, handleStsError(err, l, stsName, stsNamespace)
			}

			// Wait until verificationFunction verifies the update, passing in
			// the current partition so the function knows which pod to check
			// the status of.
			l.Info("waiting until partition done updating", "partition number:", partition)
			if err := waitUntilPerPodVerificationFuncVerifies(updateSts, perPodVerificationFunc, int(partition), updateTimer, l); err != nil {
				return false, errors.Wrapf(err, "error while running verificationFunc on pod %d", int(partition))
			}
			if partition > 0 {
				// Must refresh STS object, or the next time through the loop
				// Kubernetes will error out because the object has been updated
				// since we last read it.
				var err error
				sts, err = updateSts.clientset.AppsV1().StatefulSets(stsNamespace).Get(updateSts.ctx, stsName, metav1.GetOptions{})
				if err != nil {
					return false, handleStsError(err, l, stsName, stsNamespace)
				}
				updateTimer.sleeper.Sleep(l, "between updating pods")
			}
		}
		return skipSleep, nil
	}
}

func waitUntilPerPodVerificationFuncVerifies(
	updateSts *UpdateSts,
	perPodVerificationFunc func(*UpdateSts, int, logr.Logger) error,
	podNumber int,
	updateTimer *UpdateTimer,
	l logr.Logger,
) error {
	f := func() error {
		err := perPodVerificationFunc(updateSts, podNumber, l)
		return err
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = updateTimer.podUpdateTimeout
	b.MaxInterval = updateTimer.podMaxPollingInterval
	return backoff.Retry(f, b)
}

// TODO this code might not be used.

// waitUntilAllPodsReadyInAllClusters waits until all pods in all statefulsets are in the
// ready state. The ready state implies all nodes are passing node liveness.
func makeWaitUntilAllPodsReadyFuncInAllClusters(
	numCRDBPods int,
	clientsets map[string]kubernetes.Interface,
	podUpdateTimeout time.Duration,
	maxPodPollingInterval time.Duration,
	stsName string,
) func(ctx context.Context, l logr.Logger) error {
	return func(ctx context.Context, l logr.Logger) error {

		l.Info("waiting until all pods are in the ready state")
		f := func() error {
			got := 0
			for ns, clientset := range clientsets {
				sts, err := clientset.AppsV1().StatefulSets(ns).Get(ctx, stsName, metav1.GetOptions{})
				if err != nil {
					return handleStsError(err, l, stsName, ns)
				}
				got += int(sts.Status.ReadyReplicas)
			}

			if got != numCRDBPods {
				err := fmt.Errorf("number of ready replicas is %v, not equal to num CRDB pods %v", got, numCRDBPods)
				l.Error(err, "number of ready replicas is,  not equal to num CRDB pods")
				return err
			}
			l.Info("all replicas are ready")
			return nil
		}

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = podUpdateTimeout
		b.MaxInterval = maxPodPollingInterval
		return backoff.Retry(f, b)
	}
}

// TODO there are ALOT more reason codes in k8sErrors, should we test them all?

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
