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
	"errors"
	"fmt"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// makeUpdateCockroachVersionFunction takes a cockroachImage string and returns
// a function which takes a statefulset and returns the same statefulset, with
// the CockroachDB container image within changed to the new cockroachImage.
func makeUpdateCockroachVersionFunction(
	cockroachImage, version, oldVersion string,
) func(sts *v1.StatefulSet) (*v1.StatefulSet, error) {
	return func(sts *v1.StatefulSet) (*v1.StatefulSet, error) {
		timeNow := metav1.Now()
		if val, ok := sts.Annotations[resource.CrdbHistoryAnnotation]; !ok {
			sts.Annotations[resource.CrdbHistoryAnnotation] = fmt.Sprintf("%s=%s", timeNow.Format(time.RFC3339), oldVersion)
		} else {
			sts.Annotations[resource.CrdbHistoryAnnotation] = fmt.Sprintf("%s %s=%s", val, timeNow.Format(time.RFC3339), oldVersion)
		}
		sts.Annotations[resource.CrdbVersionAnnotation] = version
		sts.Annotations[resource.CrdbContainerImageAnnotation] = cockroachImage
		for i := range sts.Spec.Template.Spec.Containers {
			container := &sts.Spec.Template.Spec.Containers[i]
			// TODO "db" is hardcoded.  Make this a value in statefulset resource
			// so that we are sharing values here
			if container.Name == "db" {
				container.Image = cockroachImage
				return sts, nil
			}
		}
		return nil, errors.New("cockroachdb container not found in sts")
	}
}

// TODO some of these should probably be panics or cancel the update at least
// If we cannot find the Pod, or if we cannot find the container.
// We need to return a status and an error instead of just an error.

// makeWaitUntilCRDBPodIsRunningNewVersionFunction takes a cockroachImage and
// returns a function which takes a Kubernetes clientset, a statefulset, and a
// pod number within the statefulset. This function checks if the specified
// pod is running the new cockroachImage version and is in a `ready` state.
func makeIsCRBPodIsRunningNewVersionFunction(
	cockroachImage string,
) func(update *UpdateSts, podNumber int, l logr.Logger) error {
	return func(update *UpdateSts, podNumber int, l logr.Logger) error {
		sts := update.sts
		stsName := sts.Name
		stsNamespace := sts.Namespace
		podName := fmt.Sprintf("%s-%d", stsName, podNumber)
		clientset := update.clientset

		crdbPod, err := clientset.CoreV1().Pods(stsNamespace).Get(update.ctx, podName, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) { // this is not an error
			l.V(int(zapcore.DebugLevel)).Info("cannot find Pod", "podName", podName, "namespace", stsNamespace)
			return err
		} else if statusError, isStatus := err.(*k8sErrors.StatusError); isStatus { // this is an error
			l.Error(statusError, fmt.Sprintf("status error getting pod %v", statusError.ErrStatus.Message))
			return statusError
		} else if err != nil { // this is an error
			l.Error(err, "error getting pod")
			return err
		}

		for i := range crdbPod.Spec.Containers {
			container := &crdbPod.Spec.Containers[i]
			// TODO this is hard coded and resource statefulset needs to use this
			if container.Name == "db" {

				// TODO this is not an error but should return a wait status
				if container.Image != cockroachImage {
					l.V(int(zapcore.DebugLevel)).Info("Pod is not updated to current image.")
					return fmt.Errorf("%s pod is on image %s, expected %s", podName, container.Image, cockroachImage)
				}

				// TODO this is not an error but should return a wait status
				// CRDB pod is updated to new Cockroach image. Now check
				// that the pod is in a ready state before proceeding.
				if !kube.IsPodReady(crdbPod) {
					l.V(int(zapcore.DebugLevel)).Info("Pod is not ready yet.", "podName", podName, "stsName", stsName)
					return fmt.Errorf("%s pod not ready yet", crdbPod)
				}

				l.V(int(zapcore.DebugLevel)).Info("is running new version on", "podName", podName, "stsName", stsNamespace)
				return nil
			}
		}

		// TODO how do we even get here??
		// I am not certain that this code is even used.

		err = fmt.Errorf("cockroachdb container not found within the cockroach pod")
		l.Error(err, "container not found")
		return err
	}
}

// Note that while CockroachDB considers 19.2 to be a major release, if we follow
// semantic versioning (https://semver.org/spec/v2.0.0.html), both 19.1 and 19.2
// is a minor release of version 19. The code below parses the version as if it
// were following the semantic versioning spec.
func isPatch(wantVersion *semver.Version, currentVersion *semver.Version) bool {
	return currentVersion.Major() == wantVersion.Major() && currentVersion.Minor() == wantVersion.Minor()
}

func isForwardOneMajorVersion(wantVersion *semver.Version, currentVersion *semver.Version) bool {
	// Two cases:
	// 19.1 to 19.2 -> same year
	// 19.2 to 20.1 -> next year
	return (currentVersion.Major() == wantVersion.Major() && currentVersion.Minor()+1 == wantVersion.Minor()) ||
		(currentVersion.Major()+1 == wantVersion.Major() && currentVersion.Minor()-1 == wantVersion.Minor())
}

func isBackOneMajorVersion(wantVersion *semver.Version, currentVersion *semver.Version) bool {
	// Two cases:
	// 19.2 to 19.1 -> same year
	// 20.1 to 19.2 -> previous year
	return (currentVersion.Major() == wantVersion.Major() && currentVersion.Minor() == wantVersion.Minor()+1) ||
		(currentVersion.Major() == wantVersion.Major()+1 && currentVersion.Minor() == wantVersion.Minor()-1)
}
