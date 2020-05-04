package kube

import (
	"bytes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func ObjectKey(obj runtime.Object) types.NamespacedName {
	accessor, _ := meta.Accessor(obj)
	return types.NamespacedName{
		Namespace: accessor.GetNamespace(),
		Name:      accessor.GetName(),
	}
}

func ExecInPod(scheme *runtime.Scheme, config *rest.Config, namespace string, name string, container string, cmd []string) (string, string, error) {
	tty := false
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to create kubernetes clientset")
	}

	req := client.CoreV1().RESTClient().Post().
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
