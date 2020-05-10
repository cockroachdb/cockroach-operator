package actor

import (
	"context"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/condition"
	"github.com/cockroachlabs/crdb-operator/pkg/kube"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newDeploy(scheme *runtime.Scheme, cl client.Client) Actor {
	return &deploy{
		action: newAction("deploy", scheme, cl),
	}
}

type deploy struct {
	action
}

func (d deploy) Handles(conds []api.ClusterCondition) bool {
	return condition.False(api.NotInitializedCondition, conds) || condition.True(api.NotInitializedCondition, conds)
}

func (d deploy) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := d.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.Info("reconciling resources")

	r := resource.NewManagedKubeResource(ctx, d.client, cluster, kube.AnnotatingPersister)

	owner := cluster.Unwrap()

	changed, err := (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.DiscoveryServiceBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(),
		},
		Owner:  owner,
		Scheme: d.scheme,
	}).Reconcile()
	if err != nil {
		return errors.Wrap(err, "failed to reconcile discovery service")
	}

	if changed {
		log.Info("created/updated discovery service, stopping request processing")
		CancelLoop(ctx)
		return nil
	}

	changed, err = (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.PublicServiceBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(),
		},
		Owner:  owner,
		Scheme: d.scheme,
	}).Reconcile()
	if err != nil {
		return errors.Wrap(err, "failed to reconcile public service")
	}

	if changed {
		log.Info("created/updated public service, stopping request processing")
		CancelLoop(ctx)
		return nil
	}

	changed, err = (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.StatefulSetBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(),
		},
		Owner:  owner,
		Scheme: d.scheme,
	}).Reconcile()
	if err != nil {
		return errors.Wrap(err, "failed to reconcile statefulset")
	}

	if changed {
		log.Info("created/updated statefulset, stopping request processing")
		CancelLoop(ctx)
		return nil
	}

	log.Info("completed")
	return nil
}
