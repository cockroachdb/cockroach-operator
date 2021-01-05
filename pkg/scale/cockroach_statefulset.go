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

package scale

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cockroachdb/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	autoscaling "k8s.io/api/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
)

//  backoffFactory is a replacable global for backoff creation. It may be
// replaced with shorter times to allow testing of Wait___ functions without
// waiting the entire default period
var backoffFactory = defaultBackoffFactory

//ClusterScaler interface
type ClusterScaler interface {
	Replicas(context.Context) (uint, error)
	SetReplicas(context.Context, uint) error
	WaitUntilRunning(context.Context) error
	WaitUntilHealthy(context.Context, uint) error
}

// CockroachStatefulSet represents the CRDB statefulset running in a kubernetes cluster
type CockroachStatefulSet struct {
	Name      string
	Namespace string
	ClientSet kubernetes.Interface
}

func defaultBackoffFactory(maxTime time.Duration) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxTime
	return b
}

// Replicas returns the number of desired replicas for CRDB's statefulset
func (c *CockroachStatefulSet) Replicas(ctx context.Context) (uint, error) {
	// Note: GetScale would be better here but k8s fake client set doesn't play nice with it...
	// This is effectively the same operation.
	sst, err := c.ClientSet.AppsV1().StatefulSets(c.Namespace).Get(ctx, c.Name, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	if sst.Spec.Replicas == nil {
		return 0, errors.New("replicas not set on .Spec")
	}

	return uint(*sst.Spec.Replicas), nil
}

// SetReplicas sets the desired replicas for CRDB's statefulset without waiting for
// new pods to be created or to become healthy.
func (c *CockroachStatefulSet) SetReplicas(ctx context.Context, scale uint) error {
	_, err := c.ClientSet.AppsV1().StatefulSets(c.Namespace).UpdateScale(ctx, c.Name, &autoscaling.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
		Spec: autoscaling.ScaleSpec{
			Replicas: int32(scale),
		},
	}, metav1.UpdateOptions{})

	return errors.Wrap(err, "failed to update statefulset scale")
}

// WaitUntilRunning blocks until the target statefulset has the expected number of pods running but not necessarily ready
func (c *CockroachStatefulSet) WaitUntilRunning(ctx context.Context) error {
	return WaitUntilStatefulSetIsRunning(ctx, c.ClientSet, c.Namespace, c.Name)
}

// WaitUntilHealthy blocks until the target stateful set has exactly `scale` healthy replicas.
func (c *CockroachStatefulSet) WaitUntilHealthy(ctx context.Context, scale uint) error {
	return WaitUntilStatefulSetIsReadyToServe(ctx, c.ClientSet, c.Namespace, c.Name, int32(scale))
}

// WaitUntilStatefulSetIsRunning waits until the given statefulset has all pods scheduled and running but not necessarily healthy nor ready
func WaitUntilStatefulSetIsRunning(ctx context.Context, clientset kubernetes.Interface, namespace string, name string) error {

	f := func() error {
		return StatefulSetIsRunning(ctx, clientset, namespace, name)
	}

	b := backoffFactory(5 * time.Minute)
	b = backoff.WithContext(b, ctx)

	if err := backoff.Retry(f, b); err != nil {
		return errors.Wrapf(err, "statefulSet is not running: %s", name)
	}

	return nil
}

// StatefulSetIsRunning checks if the expected number of pods for a statefulset are running but not necessarily ready nor healthy
func StatefulSetIsRunning(ctx context.Context, clientset kubernetes.Interface, namespace string, name string) error {
	sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get statefulset: %s", name)
	}

	if sts.Spec.Replicas == nil {
		return errors.New("statefulset has nil replicas")
	}

	if sts.Status.Replicas != *sts.Spec.Replicas {
		return errors.Errorf(
			"statefulset replicas not yet reconciled. have %d expected %d",
			sts.Status.Replicas,
			sts.Spec.Replicas,
		)
	}

	pods, err := clientset.CoreV1().Pods(sts.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(sts.Spec.Selector.MatchLabels).AsSelector().String(),
		FieldSelector: "status.phase=Running",
	})

	if err != nil {
		return errors.Wrapf(err, "failed to list statefulset pods")
	}

	if int32(len(pods.Items)) != *sts.Spec.Replicas {
		return errors.Errorf("not all statefulset pods are running")
	}

	return nil
}

//WaitUntilStatefulSetIsReadyToServe func
func WaitUntilStatefulSetIsReadyToServe(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace, name string,
	numReplicas int32) error {

	f := func() error {
		return IsStatefulSetReadyToServe(ctx, clientset, namespace, name, numReplicas)
	}

	b := backoffFactory(5 * time.Minute)
	return backoff.Retry(f, backoff.WithContext(b, ctx))
}

//IsStatefulSetReadyToServe func
func IsStatefulSetReadyToServe(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace, name string,
	numReplicas int32,
) error {
	ss, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "fetching cockroachdb statefulset")
	}
	if ss.Status.ReadyReplicas != numReplicas {
		return errors.Newf("all cockroachdb pods in %s are not ready yet", namespace)
	}
	if ss.Status.Replicas > numReplicas {
		return errors.Newf("not all extra replicas in %s are terminated", namespace)
	}
	return nil
}
