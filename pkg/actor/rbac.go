package actor

import (
	"context"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/condition"
	"github.com/cockroachlabs/crdb-operator/pkg/kube"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// appsv1 "k8s.io/api/apps/v1"
	// kubetypes "k8s.io/apimachinery/pkg/types"
)

type rbac struct {
	action

	config *rest.Config
}

func newRBAC(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &rbac{
		action: newAction("rbac", scheme, cl),
		config: config,
	}
}

func (r rbac) Handles(conds []api.ClusterCondition) bool {
	return condition.True(api.NotInitializedCondition, conds)
}

func (rbac rbac) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := rbac.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.Info("initializing CockroachDB RBAC")

	r := resource.NewManagedKubeResource(ctx, rbac.client, cluster, kube.AnnotatingPersister)

	owner := cluster.Unwrap()

	changed, err := (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.ServiceAccountBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(),
		},
		Owner:  owner,
		Scheme: rbac.scheme,
	}).Reconcile()
	if err != nil {
		return errors.Wrap(err, "failed to reconcile service account")
	}

	if changed {
		log.Info("created/updated service account, stopping request processing")
		// CancelLoop(ctx)
		// return nil
	}

	changed, err = (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.ClusterRoleBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(),
		},
		Owner:  owner,
		Scheme: rbac.scheme,
	}).Reconcile()
	if err != nil {
		return errors.Wrap(err, "failed to reconcile cluster role")
	}

	if changed {
		log.Info("created/updated cluster role, stopping request processing")
		// CancelLoop(ctx)
		// return nil
	}

	changed, err = (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.ClusterRoleBindingBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(),
		},
		Owner:  owner,
		Scheme: rbac.scheme,
	}).Reconcile()
	if err != nil {
		return errors.Wrap(err, "failed to reconcile cluster role binding")
	}

	if changed {
		log.Info("created/updated cluster role binding, stopping request processing")
		// CancelLoop(ctx)
		// return nil
	}

	log.Info("completed")
	return nil
}
