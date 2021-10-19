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

package kube

import (
	"bytes"
	"context"
	"fmt"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/cenkalti/backoff"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	json "github.com/json-iterator/go"
)

const LastAppliedAnnotation = "crdb.io/last-applied"

var annotator = patch.NewAnnotator(LastAppliedAnnotation)
var patchMaker = patch.NewPatchMaker(annotator)

func ExecInPod(scheme *runtime.Scheme, config *rest.Config, namespace string, name string, container string, cmd []string) (string, string, error) {
	tty := false
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to create kubernetes clientset")
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec")

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Command:   cmd,
		Container: container,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       tty,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to initialize SPDY executor")
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    tty,
	})
	if err != nil {
		return "", stderr.String(), errors.Wrapf(err, "failed to stream execution results back")
	}

	return stdout.String(), stderr.String(), nil
}

func GetClusterCA(ctx context.Context, config *rest.Config) ([]byte, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create kubernetes clientset")
	}

	cm, err := clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "extension-apiserver-authentication", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch config map with cluster CA")
	}

	if bundle, ok := cm.Data["client-ca-file"]; ok {
		return []byte(bundle), nil
	}

	return nil, errors.New("no cluster CA found")
}

func Get(ctx context.Context, cl client.Client, obj client.Object) error {
	key := client.ObjectKeyFromObject(obj)

	return cl.Get(ctx, key, obj)
}

func FindContainer(container string, spec *corev1.PodSpec) (*corev1.Container, error) {
	for i := range spec.Containers {
		if spec.Containers[i].Name == container {
			return &spec.Containers[i], nil
		}
	}

	return nil, fmt.Errorf("failed to find container %s", container)
}

type PersistFn func(context.Context, client.Client, client.Object, MutateFn) (upserted bool, err error)

var DefaultPersister PersistFn = func(ctx context.Context, cl client.Client, obj client.Object, f MutateFn) (upserted bool, err error) {
	result, err := ctrl.CreateOrUpdate(ctx, cl, obj, func() error {
		return f()
	})

	return result == ctrlutil.OperationResultCreated || result == ctrlutil.OperationResultUpdated, err
}

var AnnotatingPersister PersistFn = func(ctx context.Context, cl client.Client, obj client.Object, f MutateFn) (upserted bool, err error) {
	return CreateOrUpdateAnnotated(ctx, cl, obj, func() error {
		return f()
	})
}

// MutateFn is a function which mutates the existing object into it's desired state.
type MutateFn func() error

func CreateOrUpdateAnnotated(ctx context.Context, c client.Client, obj client.Object, f MutateFn) (upserted bool, err error) {
	key := client.ObjectKeyFromObject(obj)

	if err := c.Get(ctx, key, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, err
		}

		if err := mutate(f, key, obj); err != nil {
			return false, err
		}

		if err := annotator.SetLastAppliedAnnotation(obj); err != nil {
			return false, err
		}

		if err := c.Create(ctx, obj); err != nil {
			return false, err
		}

		return true, nil
	}

	existing := obj.DeepCopyObject()
	if err := mutate(f, key, obj); err != nil {
		return false, err
	}

	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}

	switch obj.(type) {
	case *appsv1.StatefulSet:
		opts = append(opts, IgnoreVolumeClaimTemplatesAndMode(), patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus())
	}

	patchResult, err := patchMaker.Calculate(existing, obj, opts...)
	if err != nil {
		return false, err
	}

	if patchResult.IsEmpty() {
		return false, nil
	}

	if err := annotator.SetLastAppliedAnnotation(obj); err != nil {
		return false, err
	}

	if err := c.Update(ctx, obj); err != nil {
		return false, err
	}
	return true, nil
}

func IgnoreVolumeClaimTemplatesAndMode() patch.CalculateOption {
	return func(current, modified []byte) ([]byte, []byte, error) {
		current, err := pruneVolumeInformation(current)
		if err != nil {
			return []byte{}, []byte{}, errors.Wrap(err, "could not delete volumeMode field from current byte sequence")
		}

		modified, err = pruneVolumeInformation(modified)
		if err != nil {
			return []byte{}, []byte{}, errors.Wrap(err, "could not delete volumeMode field from modified byte sequence")
		}

		return current, modified, nil
	}
}

func pruneVolumeInformation(obj []byte) ([]byte, error) {
	resource := map[string]interface{}{}
	err := json.Unmarshal(obj, &resource)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not unmarshal byte sequence")
	}

	if spec, ok := resource["spec"]; ok {
		if spec, ok := spec.(map[string]interface{}); ok {
			if vctl, ok := spec["volumeClaimTemplates"]; ok {
				if vcto, ok := vctl.([]interface{}); ok {
					for _, vct := range vcto {
						if vct, ok := vct.(map[string]interface{}); ok {
							if vcts, ok := vct["spec"].(map[string]interface{}); ok {
								vcts["volumeMode"] = "Filesystem"
							}
						}
					}
				}
			}
		}
	}

	obj, err = json.ConfigCompatibleWithStandardLibrary.Marshal(resource)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not marshal byte sequence")
	}

	return obj, nil
}

func mutate(f MutateFn, key client.ObjectKey, obj client.Object) error {
	if err := f(); err != nil {
		return err
	}

	newKey := client.ObjectKeyFromObject(obj)
	if key.String() != newKey.String() {
		return fmt.Errorf("MutateFn cannot mutate object name and/or object namespace")
	}

	return nil
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

//IsImagePullBackOff  returns true if a container status has the waiting state with reason ImagePullBackOff
func IsImagePullBackOff(pod *corev1.Pod, image string) bool {
	_, containerStatus := GetContainerStatus(&pod.Status, image)
	if containerStatus != nil && !containerStatus.Ready && containerStatus.State.Waiting != nil &&
		containerStatus.State.Waiting.Reason == "ImagePullBackOff" {
		return true
	}
	return false
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

// GetContainerStatus extracts the provided container status from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located container status.
func GetContainerStatus(status *corev1.PodStatus, image string) (int, *corev1.ContainerStatus) {
	if status == nil {
		return -1, nil
	}
	return GeContainerStatusFromList(status.ContainerStatuses, image)
}

// GeContainerStatusFromList extracts the provided container status from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the containeer status is not present.
func GeContainerStatusFromList(containerStatuses []corev1.ContainerStatus, image string) (int, *corev1.ContainerStatus) {
	if containerStatuses == nil {
		return -1, nil
	}
	for i := range containerStatuses {
		if containerStatuses[i].Image == image {
			return i, &containerStatuses[i]
		}
	}
	return -1, nil
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

// WaitUntilAllStsPodsAreReady waits until all pods in the statefulset are in the
// ready state. The ready state implies all nodes are passing node liveness.
func WaitUntilAllStsPodsAreReady(ctx context.Context, clientset *kubernetes.Clientset, l logr.Logger, stsname, stsnamespace string, podUpdateTimeout, podMaxPollingInterval time.Duration) error {
	l.V(int(zapcore.DebugLevel)).Info("waiting until all pods are in the ready state")
	f := func() error {
		sts, err := clientset.AppsV1().StatefulSets(stsnamespace).Get(ctx, stsname, metav1.GetOptions{})
		if err != nil {
			return HandleStsError(err, l, stsname, stsnamespace)
		}
		got := int(sts.Status.ReadyReplicas)
		// TODO need to test this
		// we could also use the number of pods defined by the operator
		numCRDBPods := int(sts.Status.Replicas)
		if got != numCRDBPods {
			l.Error(err, fmt.Sprintf("number of ready replicas is %v, not equal to num CRDB pods %v", got, numCRDBPods))
			return err
		}

		l.V(int(zapcore.DebugLevel)).Info("all replicas are ready makeWaitUntilAllPodsReadyFunc update_cockroach_version.go")
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = podUpdateTimeout
	b.MaxInterval = podMaxPollingInterval
	return backoff.Retry(f, b)
}

func HandleStsError(err error, l logr.Logger, stsName string, ns string) error {
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

// MergeAnnotations merges the `from` annotations into `to` annotations
func MergeAnnotations(to, from map[string]string) {
	for key, value := range from {
		to[key] = value
	}
}
