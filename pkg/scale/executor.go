package scale

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/cockroachdb/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type CockroachExecutor struct {
	Namespace string
	StatefulSet string
	Container string
	Config    *rest.Config
	ClientSet kubernetes.Interface
	TTY       bool
}

func (e CockroachExecutor) Exec(ctx context.Context, podIdx uint, cmd []string) (string, string, error) {
	var stdout, stderr bytes.Buffer

	err := Executor{Namespace: e.Namespace, Config: e.Config}.Exec(ctx, ExecutorOptions{
		Pod:       fmt.Sprintf("%s-%d", e.StatefulSet, podIdx),
		Container: e.Container,
		Cmd:       cmd,
		Stdout:    &stdout,
		Stderr:    &stderr,
		TTY:       e.TTY,
	})

	if err != nil {
		return "", stderr.String(), errors.Wrapf(err, "failed to stream execution results back")
	}

	return stdout.String(), stderr.String(), nil
}

type Executor struct {
	Namespace string
	Config    *rest.Config
}

type ExecutorOptions struct {
	Pod       string
	Container string
	Cmd       []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	TTY       bool
}

func (e Executor) Exec(ctx context.Context, o ExecutorOptions) error {
	cs, err := kubernetes.NewForConfig(e.Config)
	if err != nil {
		return err
	}

	req := cs.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(o.Pod).
		Namespace(e.Namespace).
		SubResource("exec")

	// TODO(chrisseto): How do we get context support here? Previously there
	// was a .Context(ctx) call on the request object

	req.VersionedParams(&corev1.PodExecOptions{
		Command:   o.Cmd,
		Container: o.Container,
		Stdin:     o.Stdin != nil,
		Stdout:    o.Stdout != nil,
		Stderr:    o.Stderr != nil,
		TTY:       o.TTY,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(e.Config, "POST", req.URL())
	if err != nil {
		return errors.Wrapf(err, "failed to initialize SPDY executor")
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  o.Stdin,
		Stdout: o.Stdout,
		Stderr: o.Stderr,
		Tty:    false,
	})
}
