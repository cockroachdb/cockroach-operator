package update

import (
	"context"
	"fmt"

	semver "github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// makeUpdateCockroachVersionFunction takes a cockroachImage string and returns
// a function which takes a statefulset and returns the same statefulset, with
// the CockroachDB container image within changed to the new cockroachImage.
func makeUpdateCockroachVersionFunction(
	cockroachImage string,
) func(sts *v1.StatefulSet) (*v1.StatefulSet, error) {
	return func(sts *v1.StatefulSet) (*v1.StatefulSet, error) {
		for i := range sts.Spec.Template.Spec.Containers {
			container := &sts.Spec.Template.Spec.Containers[i]
			if container.Name == "cockroachdb" {
				container.Image = cockroachImage
				return sts, nil
			}
		}
		return nil, errors.New("cockroachdb container not found in sts")
	}
}

// makeWaitUntilCRDBPodIsRunningNewVersionFunction takes a cockroachImage and
// returns a function which takes a Kubernetes clientset, a statefulset, and a
// pod number within the statefulset. This function checks if the specified
// pod is running the new cockroachImage version and is in a `ready` state.
func makeIsCRBPodIsRunningNewVersionFunction(
	cockroachImage string,
) func(update *UpdateSts, podNumber int, l *zap.Logger) error {
	return func(update *UpdateSts, podNumber int, l *zap.Logger) error {
		sts := update.sts
		stsName := sts.Name
		stsNamespace := sts.Namespace
		podName := fmt.Sprintf("%s-%d", stsName, podNumber)
		clientset := update.clientset
		crdbPod, err := clientset.CoreV1().Pods(stsNamespace).Get(context.TODO(), stsName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "error getting cockroach pod")
		}
		for i := range crdbPod.Spec.Containers {
			container := &crdbPod.Spec.Containers[i]
			if container.Name == "cockroachdb" {
				if container.Image != cockroachImage {
					return fmt.Errorf("%s pod is on image %s, expected %s", podName, container.Image, cockroachImage)
				}
				// CRDB pod is updated to new Cockroach image. Now check
				// that the pod is in a ready state before proceeding.
				if !IsPodReady(crdbPod) {
					return fmt.Errorf("%s pod not ready yet", podName)
				}
				l.Sugar().Infof("%s is running new version on %s", podName, stsNamespace)
				return nil
			}
		}
		return fmt.Errorf("cockroachdb container not found within the cockroach pod")
	}
}

// TODO this code is from https://github.com/kubernetes/kubernetes/blob/master/pkg/api/v1/pod/util.go
// We need to determine if this functionality is available via the client-go

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *corev1.Pod) bool {
	return IsPodReadyConditionTrue(pod.Status)
}

// IsPodReadyConditionTrue returns true if a pod is ready; false otherwise.
func IsPodReadyConditionTrue(status corev1.PodStatus) bool {
	condition := GetPodReadyCondition(status)
	return condition != nil && condition.Status == corev1.ConditionTrue
}

// GetPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func GetPodReadyCondition(status corev1.PodStatus) *corev1.PodCondition {
	_, condition := GetPodCondition(&status, corev1.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetPodCondition(status *corev1.PodStatus, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return GetPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetPodConditionFromList(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) (int, *corev1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
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
