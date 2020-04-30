package actor

import (
	"context"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/condition"
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
	return condition.False(api.InitializedCondition, conds)
}

func (d deploy) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := d.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.Info("reconciling resources")

	r := resource.NewManagedKubeResource(ctx, d.client, cluster)

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

	planner := TopologyPlanner{
		Cluster: cluster,
	}

	err = planner.ForEachZone(func(name string, nodesNum int32, join string, locality string, nodeSelector map[string]string) error {
		// Skip if one statefulset was updated to reschedule reconciliation
		if changed {
			return nil
		}

		changed, err = (resource.Reconciler{
			ManagedResource: r,
			Builder:         resource.NewStatefulSetBuilder(cluster, name, nodesNum, join, locality, nodeSelector),
			Owner:           owner,
			Scheme:          d.scheme,
		}).Reconcile()
		if err != nil {
			return err
		}

		return nil
	})

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
