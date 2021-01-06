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

package resource

import (
	"context"

	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Builder populates a given Kubernetes resource or creates its default instance (placeholder)
type Builder interface {
	Build(runtime.Object) error
	Placeholder() runtime.Object
}

// Fetcher updates the object with its state from Kubernetes
type Fetcher interface {
	Fetch(obj runtime.Object) error
}

// Persister creates or updates the object in Kubernetes after calling the mutation function.
type Persister interface {
	Persist(obj runtime.Object, mutateFn func() error) (upserted bool, err error)
}

func NewKubeResource(ctx context.Context, client client.Client, namespace string, persistFn kube.PersistFn) Resource {
	return Resource{
		Fetcher:   NewKubeFetcher(ctx, namespace, client),
		Persister: NewKubePersister(ctx, namespace, client, persistFn),
	}
}

// Resource represents a resource that can be fetched or saved
type Resource struct {
	Fetcher
	Persister
}

func NewManagedKubeResource(ctx context.Context, client client.Client, cluster *Cluster, persistFn kube.PersistFn) ManagedResource {
	return ManagedResource{
		Resource: NewKubeResource(ctx, client, cluster.Namespace(), persistFn),

		Labels: labels.Common(cluster.Unwrap()),
	}
}

// ManagedResource is a `Resource` with labels which can be reconciled by `Reconciler`
type ManagedResource struct {
	Resource

	labels.Labels
}

// Reconciler reconciles managed Kubernetes resource with `Builder` results
type Reconciler struct {
	ManagedResource

	Builder
	Owner  metav1.Object
	Scheme *runtime.Scheme
}

func (r Reconciler) Reconcile() (upserted bool, err error) {
	current := r.Placeholder()

	if err := r.Fetch(current); kube.IgnoreNotFound(err) != nil {
		return false, err
	}

	original := current.DeepCopyObject()

	return r.Persist(current, func() error {
		if err := r.Build(current); err != nil {
			return err
		}

		if err := r.reconcileLabels(original, current); err != nil {
			return errors.Wrap(err, "failed to reconcile labels")
		}

		if err := r.reconcileAnnotations(original, current); err != nil {
			return errors.Wrap(err, "failed to reconcile annotations")
		}

		if err := r.ensureIsOwned(current); err != nil {
			return errors.Wrap(err, "failed to set object ownership")
		}

		return nil
	})
}

func (r Reconciler) reconcileLabels(current, desired runtime.Object) error {
	ll, err := labels.FromObject(current)
	if err != nil {
		return err
	}

	labels.Update(ll, r.Labels)

	return ll.ApplyTo(desired)
}

func (r Reconciler) reconcileAnnotations(current, desired runtime.Object) error {
	caccessor, err := meta.Accessor(current)
	if err != nil {
		return err
	}

	if caccessor.GetAnnotations() == nil {
		return nil
	}

	daccessor, err := meta.Accessor(desired)
	if err != nil {
		return err
	}

	aa := daccessor.GetAnnotations()
	if aa == nil {
		aa = make(map[string]string)
	}

	for k, v := range caccessor.GetAnnotations() {
		_, found := aa[k]
		if found {
			continue
		}
		aa[k] = v
	}

	return nil
}

func (r Reconciler) ensureIsOwned(desired runtime.Object) error {
	metaObj, err := meta.Accessor(desired)
	if err != nil {
		return err
	}

	return ctrl.SetControllerReference(r.Owner, metaObj, r.Scheme)
}

func NewKubeFetcher(ctx context.Context, namespace string, reader client.Reader) *KubeFetcher {
	return &KubeFetcher{
		ctx:       ctx,
		namespace: namespace,
		Reader:    reader,
	}
}

// KubeFetcher fetches Kubernetes results
type KubeFetcher struct {
	ctx       context.Context
	namespace string
	client.Reader
}

func (f KubeFetcher) Fetch(o runtime.Object) error {
	accessor, err := meta.Accessor(o)
	if err != nil {
		return err
	}

	err = f.Reader.Get(f.ctx, f.makeKey(accessor.GetName()), o)

	return err
}

func (f KubeFetcher) makeKey(name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: f.namespace,
	}
}

func NewKubePersister(ctx context.Context, namespace string, client client.Client, persistFn kube.PersistFn) *KubePersister {
	return &KubePersister{
		ctx:       ctx,
		namespace: namespace,
		persistFn: persistFn,
		Client:    client,
	}
}

// KubePersister saves resources back into Kubernetes
type KubePersister struct {
	ctx       context.Context
	namespace string
	persistFn kube.PersistFn
	client.Client
}

func (p KubePersister) Persist(obj runtime.Object, mutateFn func() error) (upserted bool, err error) {
	if err := addNamespace(obj, p.namespace); err != nil {
		return false, err
	}

	return p.persistFn(p.ctx, p.Client, obj, mutateFn)
}

func addNamespace(o runtime.Object, ns string) error {
	accessor, err := meta.Accessor(o)
	if err != nil {
		return errors.Wrapf(err, "failed to access object's meta information")
	}

	accessor.SetNamespace(ns)

	return nil
}
