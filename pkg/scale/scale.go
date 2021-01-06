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

	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
)

//PVCPruner interface
type PVCPruner interface {
	Prune(ctx context.Context) error
}

//Scaler interface
type Scaler struct {
	Logger    logr.Logger
	CRDB      ClusterScaler
	Drainer   Drainer
	PVCPruner PVCPruner
}

// EnsureScale gracefully adds or removes CRDB replicas from a given stateful
// until it matches the given number. Removed nodes are drained of all replicas
// before being removed from the CRDB cluster and their matching PVCs and PVs
// will be removed as well.
// In some cases, it may not be possible to full drain a node. In such cases a
// ErrDecommissioningStalled will be returned  and the node will be left in a
// decommissioning  state.
func (s *Scaler) EnsureScale(ctx context.Context, scale uint) error {
	// Before doing any scaling, prune any PVCs that are not currently in use.
	// This only needs to be done when scaling up but the operation is a noop
	// if there are no PVCs not currently in use.
	// As of v20.2.0, CRDB nodes may not be recommissioned. To account for
	// this, PVCs must be removed (pruned) before scaling up to avoid reusing a
	// previously decommissioned node.
	// Prune MUST be called before scaling as older clusters may have dangling
	// PVCs.
	// All underlying PVs and the storageclasses they were created with should
	// make use of reclaim policy = delete. A reclaim policy of retain is fine
	// but will result in wasted money, recycle should be considered unsafe and
	// is officially deprecated by kubernetes.
	if err := s.PVCPruner.Prune(ctx); err != nil {
		return errors.Wrap(err, "initial PVC pruning")
	}

	// NOTE: Scaling down 3 -> 1 does not currently work (gracefully).
	// See https://github.com/cockroachlabs/managed-service/issues/2751 for more details
	// NOTE: Scaling up 1 -> 3 will not update monitoring, so SREs will not be alerted of issues.
	crdbScale, err := s.CRDB.Replicas(ctx)
	if err != nil {
		return err
	}

	// TODO (chrisseto): To mitigate some of the issues with adding multiple clusters at a time we should
	// set kv.snapshot_rebalance.max_rate and kv.snapshot_rebalance.max_rate to ~2MB.
	// Given the low number of IOPs provisioned by CC it seems likely that 2MB would be a reasonable default
	// SET CLUSTER SETTING kv.snapshot_rebalance.max_rate='2MB'
	// SET CLUSTER SETTING kv.snapshot_recovery.max_rate='2MB'

	// Scale down CRDB if need be. Do it gracefully one by one for safety
	for crdbScale > scale {
		oneOff := crdbScale - 1

		s.Logger.Info("scaling down stateful set", "have", crdbScale, "want", oneOff)

		// TODO (chrisseto): If decommissioning fails due to a timeout
		// recommission that node before failing this job.
		// Making use of the on finish hook is likely ideal?
		if err := s.Drainer.Decommission(ctx, oneOff); err != nil {
			return err
		}

		if err := s.CRDB.SetReplicas(ctx, oneOff); err != nil {
			return err
		}

		if err := s.CRDB.WaitUntilHealthy(ctx, oneOff); err != nil {
			return err
		}

		if crdbScale, err = s.CRDB.Replicas(ctx); err != nil {
			return err
		}
	}

	// Scale up one node at a time to:
	// 1. Mitigate race conditions.
	// Some legacy clusters have a parallel pod management strategy.
	// (Pod management strategy is immutable)
	// This can lead to "unscalable" k8s topologies if many pods are created at once.
	// | A | B | C |
	// | 0 |   |   |
	// Could result in the following if 4 nodes are added at the same time
	// | A | B | C |
	// | 0 | 1 | 3 |
	// | 4 | 2 |   |
	// Scaling back to 3 nodes would then result in zone C being removed
	// The K8s scheduler, when scaling one at a time, will ensure "correct" distributions
	// | A | B | C |
	// | 0 | 1 | 2 |
	// | 3 | 2 |   |
	// 2. Minimize the impact on the cluster's performance.
	// When adding many nodes to a cluster that is under load, it is possible to max out the available IOPS.
	// This results is slower overall performance and slower integration of the new nodes in the cluster
	// and may even look like adding new nodes has resulted in worse performance for a few hours.
	// 3. Keep us a bit sane.
	// When nodes are added in order they'll gain node ids equal to their pod index + 1.
	// If all pods are created in parallel cockroachdb-3 may end up being node n1.
	//
	// Final note: It does appear possible to alter immutable fields in statefulsets by deleting them
	// with cascade = false and recreating them. They will "adopt" the old/existing pods and in theory not
	// have an affect on the cluster as a whole.
	for crdbScale < scale {
		s.Logger.Info("scaling up stateful set", "have", crdbScale, "want", (crdbScale + 1))
		if err := s.CRDB.SetReplicas(ctx, crdbScale+1); err != nil {
			return err
		}

		// Wait for the newly requested pod to be scheduled and running
		if err := s.CRDB.WaitUntilRunning(ctx); err != nil {
			return err
		}

		// Wait for the newly running pods to become healthy
		if err := s.CRDB.WaitUntilHealthy(ctx, crdbScale+1); err != nil {
			return err
		}

		// TODO wait for cluster to be rebalanced before proceeding.
		// Uncertain how to tell when a cluster is actually rebalanced.
		// 1. Check replicas per node. Possibly flaky/difficult due to zone configs
		// 2. Check for a lack of learner replicas. Ben says this is flakey.
		// 3. Look at the range distribution reports? Haven't been able to fully explore this option.

		if crdbScale, err = s.CRDB.Replicas(ctx); err != nil {
			return err
		}
	}

	// NB: We may be able to remove the scheduler entirely once this change is
	// running in production and we have access to Pod Topology Spread
	// Constraints (stable in k8s 1.19). The scheduler's main job was to ensure
	// that pods are always scheduled into the same AZ, this ensures that PVC
	// can always be mounted to the correct pod. One of the most common cases
	// of zone crossing was during a scale up operation while the PVC had
	// already been provisioned. We don't have to worry about disks crossing
	// zones if there are no disks. taps-head.gif
	// Pod Topology Spread constraints may be able  to handle the other job of
	// the scheduled, which is forcing a "good" distribution when the sts pods
	// are initially distributed across zones.
	// Testing is just a bit difficult and the overhead of the scheduler is
	// extremely low, so we're leaving it as is for now.
	// With the call to Prune at the beginning, this call is not
	// Strictly speaking required, it is just for cost savings and
	// cleanliness.
	if err := s.PVCPruner.Prune(ctx); err != nil {
		return errors.Wrap(err, "final PVC pruning")
	}

	return nil
}
