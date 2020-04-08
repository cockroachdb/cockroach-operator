package resource

import (
	"context"
	"github.com/cockroachlabs/crdb-operator/pkg/labels"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MutateFn func() error

type Builder interface {
	Build() (runtime.Object, error)
	Placeholder() runtime.Object
}

type Fetcher interface {
	Fetch(obj runtime.Object) error
}

type OperationResult string

const (
	OperationResultUnknown OperationResult = ""
	OperationResultNone    OperationResult = "unchanged"
	OperationResultCreated OperationResult = "created"
	OperationResultUpdated OperationResult = "updated"
)

func NewOperationResult(result string) OperationResult {
	switch OperationResult(result) {
	case OperationResultNone:
		return OperationResultNone
	case OperationResultCreated:
		return OperationResultCreated
	case OperationResultUpdated:
		return OperationResultUpdated
	default:
		return OperationResultUnknown
	}
}

type Persister interface {
	Persist(runtime.Object, MutateFn) (OperationResult, error)
}

func NewKubeResource(ctx context.Context, cluster *Cluster, scheme *runtime.Scheme, client client.Client) Resource {
	ns, cr := cluster.Namespace(), cluster.Unwrap()

	return Resource{
		Fetcher:   NewKubeFetcher(ctx, ns, client),
		Persister: NewKubePersister(ctx, ns, client),
		Labels:    labels.Common(cluster.Unwrap()),

		Owner:  cr,
		Scheme: scheme,
	}
}

type Resource struct {
	Fetcher
	Persister
	labels.Labels

	Owner  metav1.Object
	Scheme *runtime.Scheme
}

type Reconciler struct {
	Resource

	Builder
}

func (r Reconciler) Reconcile() (upserted bool, err error) {
	desired, err := r.Build()
	if err != nil {
		return false, err
	}

	current := r.Placeholder()

	if err := r.Fetch(current); err != nil {
		return false, err
	}

	result, err := r.Persist(desired, func() error {
		if err := r.reconcileLabels(current, desired); err != nil {
			return errors.Wrap(err, "failed to reconcile labels")
		}

		if err := r.reconcileAnnotations(current, desired); err != nil {
			return errors.Wrap(err, "failed to reconcile annotations")
		}

		if err := r.ensureIsOwned(desired); err != nil {
			return errors.Wrap(err, "failed to set object ownership")
		}

		return nil
	})

	if result == OperationResultUnknown {
		return false, errors.New("got unknowng operation result")
	}

	return result == OperationResultUpdated || result == OperationResultCreated, err
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

	return client.IgnoreNotFound(err)
}

func (f KubeFetcher) makeKey(name string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: f.namespace,
	}
}

func NewKubePersister(ctx context.Context, namespace string, client client.Client) *KubePersister {
	return &KubePersister{
		ctx:       ctx,
		namespace: namespace,
		Client:    client,
	}
}

type KubePersister struct {
	ctx       context.Context
	namespace string
	client.Client
}

func (p KubePersister) Persist(obj runtime.Object, f MutateFn) (OperationResult, error) {
	if err := addNamespace(obj, p.namespace); err != nil {
		return OperationResultNone, err
	}

	result, err := ctrl.CreateOrUpdate(p.ctx, p.Client, obj, func() error {
		return f()
	})

	return NewOperationResult(string(result)), err
}

func addNamespace(o runtime.Object, ns string) error {
	accessor, err := meta.Accessor(o)
	if err != nil {
		return errors.Wrapf(err, "failed to access object's meta information")
	}

	accessor.SetNamespace(ns)

	return nil
}
