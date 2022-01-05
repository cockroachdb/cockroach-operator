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

package update

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO this is not working.
// We need to add an annotation the statefulset with a value that will
// cause the pods to roll.
// See https://github.com/kubernetes/kubernetes/blob/bd4d197b5267ee198c1b0a070d7398f10df68c52/staging/src/k8s.io/kubectl/pkg/polymorphichelpers/objectrestarter.go#L95

// RollingRestart performs a rolling restart on a sts.
func RollingRestart(
	ctx context.Context,
	update *UpdateRoach,
	cluster *UpdateCluster,
	l logr.Logger,
) error {

	l.WithValues(
		"rolling cluster",
		update.StsName,
	)

	l.Info("starting rolling restart")

	updateFunction := makeRollingUpdateFunc()
	perPodVerificationFunction := makeRollingUpdateVerificationFunc()
	updateStrategyFunction := PartitionedRollingUpdateStrategy(
		perPodVerificationFunction,
	)

	updateSuite := &updateFunctionSuite{
		updateFunc:         updateFunction,
		updateStrategyFunc: updateStrategyFunction,
	}

	// We are using the same rolling update that we used in partitioned update for updating
	// a container.
	if err := updateClusterStatefulSets(ctx, update, cluster, updateSuite, l); err != nil {
		return errors.Wrapf(err, "error rolling sts: %s namespace: %s", update.StsName, update.StsNamespace)
	}

	l.Info("finished rolling restart")

	return nil
}

// We need to check that the patch is updated

// makeRollingUpdateVerificationFunc ensures that the pod that was restarted is in a ready state.
func makeRollingUpdateVerificationFunc() func(update *UpdateSts, podNumber int, l logr.Logger) error {
	return func(update *UpdateSts, podNumber int, l logr.Logger) error {

		// TODO refactor code and func to handle errors vs IsNotFound or !IsPodReady

		podName := fmt.Sprintf("%s-%d", update.sts.Name, podNumber)
		crdbPod, err := update.clientset.CoreV1().Pods(update.namespace).Get(update.ctx, podName, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) { // this is not an error
			l.Info("cannot find Pod", "podName", podName, "namespace", update.sts.Namespace)
			return err
		} else if statusError, isStatus := err.(*k8sErrors.StatusError); isStatus { // this is an error
			l.Error(statusError, fmt.Sprintf("status error getting pod %v", statusError.ErrStatus.Message))
			return errors.Wrap(statusError, "got status error from k8s api")
		} else if err != nil { // this is an error
			l.Error(err, "error getting pod")
			return errors.Wrap(err, "got error getting pod from k8s api")
		}

		if !kube.IsPodReady(crdbPod) {
			l.Info("Pod is not ready yet.", "pod name", crdbPod)
			return fmt.Errorf("%s pod not ready yet", crdbPod)
		}

		return nil
	}
}

// makeRollingUpdateFunc does nothing at this point.  We have this here
// in order to reuse updateClusterStatefulSets func.
func makeRollingUpdateFunc() func(sts *v1.StatefulSet) (*v1.StatefulSet, error) {
	return func(sts *v1.StatefulSet) (*v1.StatefulSet, error) {
		return sts, nil
	}
}
