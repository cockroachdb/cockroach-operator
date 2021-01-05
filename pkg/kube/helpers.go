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

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/pkg/errors"
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

func Get(ctx context.Context, cl client.Client, obj runtime.Object) error {
	key, _ := client.ObjectKeyFromObject(obj)

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

type PersistFn func(context.Context, client.Client, runtime.Object, MutateFn) (upserted bool, err error)

var DefaultPersister PersistFn = func(ctx context.Context, cl client.Client, obj runtime.Object, f MutateFn) (upserted bool, err error) {
	result, err := ctrl.CreateOrUpdate(ctx, cl, obj, func() error {
		return f()
	})

	return result == ctrlutil.OperationResultCreated || result == ctrlutil.OperationResultUpdated, err
}

var AnnotatingPersister PersistFn = func(ctx context.Context, cl client.Client, obj runtime.Object, f MutateFn) (upserted bool, err error) {
	return CreateOrUpdateAnnotated(ctx, cl, obj, func() error {
		return f()
	})
}

// MutateFn is a function which mutates the existing object into it's desired state.
type MutateFn func() error

func CreateOrUpdateAnnotated(ctx context.Context, c client.Client, obj runtime.Object, f MutateFn) (upserted bool, err error) {
	key, _ := client.ObjectKeyFromObject(obj)

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
		opts = append(opts, patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus())
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

func mutate(f MutateFn, key client.ObjectKey, obj runtime.Object) error {
	if err := f(); err != nil {
		return err
	}

	if newKey, err := client.ObjectKeyFromObject(obj); err != nil || key != newKey {
		return fmt.Errorf("MutateFn cannot mutate object name and/or object namespace")
	}

	return nil
}
