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
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// PersistentVolumePruner provides a .Prune method to remove unused statefulset
// PVC and their underlying PVs. The underlying PVs SHOULD have their reclaim
// policy set to delete.
type PersistentVolumePruner struct {
	Namespace   string
	StatefulSet string
	ClientSet   kubernetes.Interface
	Logger      logr.Logger
}

// watchStatefulset establishing a watch on the given statefulset in a
// goroutine and will call the provided cancel function whenever a modification
// to the .Spec.Replicas field OR an unexpected (non-modification) event is
// observed.
// It may be used to detect concurrent modification to a statefulset when a
// multi-step operation is taking place that depends on .Spec.Replicas staying
// the same for every step.
// watchStatefulset does not block and relies on the context being cancelled to
// prevent leakage of goroutines.
func (p *PersistentVolumePruner) watchStatefulset(
	ctx context.Context,
	cancel context.CancelFunc,
	sts *appsv1.StatefulSet,
) error {
	w, err := p.ClientSet.AppsV1().StatefulSets(p.Namespace).Watch(ctx, metav1.SingleObject(sts.ObjectMeta))
	if err != nil {
		return errors.Wrapf(err, "establishing watch on statefulset %s.%s", p.Namespace, p.StatefulSet)
	}

	p.Logger.Info("established statefulset watch", "name", p.StatefulSet, "namespace", p.Namespace)

	go func() {
		defer w.Stop()

		for {
			// First, select without our result channel as an option. if ctx is
			// cancelled and watch is closed (happens in tests mostly) we'll
			// generate some log spam on zero events.
			select {
			case <-ctx.Done():
				return
			default:
			}

			select {
			// NOTE: once cancel() has been called, we'll hit this case due to
			// the for loop, which will prevent goroutines from leaking.
			case <-ctx.Done():
				return
			case evt := <-w.ResultChan():
				switch evt.Type {
				case watch.Modified:
					if modified, ok := evt.Object.(*appsv1.StatefulSet); ok {
						// Only cancel if Replicas has changed. If an update
						// happens while pruning, it's still safe to run.
						// Technically, it's safe to continue if Replicas
						// decreases. However any change to replicas is
						// unexpected so we'll err on the side of caution for
						// now.
						if modified.Spec.Replicas == nil || *modified.Spec.Replicas != *sts.Spec.Replicas {
							cancel()
						}
					}
				default:
					// cancel on any unexpected events.
					p.Logger.Info("saw an unexpected event", "event", evt)
					cancel()
				}
			}
		}
	}()

	return nil
}

// pvcsToDelete locates all PVCs that were provisioned for the given
// statefulset but are not currently in use. Use is defined as having an
// ordinal that is less than the number of expected replica for the given
// statefulset.
func (p *PersistentVolumePruner) pvcsToDelete(ctx context.Context, sts *appsv1.StatefulSet) ([]corev1.PersistentVolumeClaim, error) {
	// K8s doesn't provide a way to tell if a PVC or PV is currently in use by
	// a pod. However, it is safe to assume that any PVCs with an ordinal great
	// than or equal to the sts' Replicas is not in use. As only pods with with
	// an ordinal < Replicas will exist. Any PVCs with an ordinal less than
	// Replicas is in use. To detect this, we build a map of PVCs that we
	// consider to be in use and skip and PVCs that it contains
	// the name of.
	prefixes := make([]string, len(sts.Spec.VolumeClaimTemplates))
	pvcsToKeep := make(map[string]bool, int(*sts.Spec.Replicas)*len(sts.Spec.VolumeClaimTemplates))
	for j, pvct := range sts.Spec.VolumeClaimTemplates {
		prefixes[j] = fmt.Sprintf("%s-%s-", pvct.Name, sts.Name)

		for i := int32(0); i < *sts.Spec.Replicas; i++ {
			name := fmt.Sprintf("%s-%s-%d", pvct.Name, sts.Name, i)
			pvcsToKeep[name] = true
		}
	}

	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, errors.Wrap(err, "converting statefulset selector to metav1 selector")
	}

	pvcs, err := p.ClientSet.CoreV1().PersistentVolumeClaims(p.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing PVCs to consider deleting")
	}

	i := 0
	for _, pvc := range pvcs.Items {
		// Don't delete PVCs that are still in use.
		if pvcsToKeep[pvc.Name] {
			continue
		}

		// Ensure that any PVC we consider deleting matches the expected naming
		// convention for PVCs managed by a statefulset.
		// <mount name>-<sts name>-<ordinal>
		matched := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(pvc.Name, prefix) {
				matched = true
				break
			}
		}

		if !matched {
			continue
		}

		pvcs.Items[i] = pvc
		i++
	}

	// Filter out any pvcs that are in use OR don't start with one of our
	// expected prefixes.
	pvcs.Items = pvcs.Items[:i]

	// Lexically sort pvcs to ensure we're deleting from lowest to highest.
	// This isn't incredibly important but may save us from some race
	// conditions. PVCs will be provisioned/reused from lowest to highest. If a
	// new replica is created while we're pruning and we can't detect it or
	// detect it fast enough, we'll _hopefully_ the requested PVCs will be
	// deleting forcing a new one to be created.
	// This is not a guarantee of any kind, just hedging our bets.
	// However, it is still unexpected that this operation would happen
	// concurrently due to our coarse grain cluster locking.
	sort.Slice(pvcs.Items, func(i, j int) bool {
		return pvcs.Items[i].Name < pvcs.Items[j].Name
	})

	return pvcs.Items, nil
}

// Prune locates and removes all PVCs that belong to a given statefulset but
// are not in use. Use is determined by the .Spec.Replicas field on the
// statefulset and the PVCs' ordinal. Prune will return an error if unexpected
// PVCs are encountered (conflicting labels) or the referenced statefulset's
// .Spec.Replicas field changes will this operation is running.
// The underlying PVs' reclaim policy should be set to delete, other options
// may result in leaking volumes which cost us money.
func (p *PersistentVolumePruner) Prune(ctx context.Context) error {
	sts, err := p.ClientSet.AppsV1().StatefulSets(p.Namespace).Get(
		ctx,
		p.StatefulSet,
		metav1.GetOptions{},
	)
	if err != nil {
		return errors.Wrapf(err, "getting statefulset %s.%s", p.Namespace, p.StatefulSet)
	}

	// Sanity that we can deference Replicas without panicking.
	if sts.Spec.Replicas == nil {
		return errors.New("statefulset had nil .Replicas")
	}

	// TODO we should pass in the controller context here
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// watchStatefulset will cancel our ctx if it detects any modifications to
	// sts.Spec.Replicas OR on any unexpected events, namely Deletions.
	if err := p.watchStatefulset(ctx, cancel, sts); err != nil {
		return errors.Wrap(err, "setting up statefulset watcher")
	}

	pvcs, err := p.pvcsToDelete(ctx, sts)
	if err != nil {
		return errors.Wrap(err, "finding pvcs to prune")
	}

	// 60 seconds was picked arbitrarily
	gracePeriod := int64(60)
	propagationPolicy := metav1.DeletePropagationForeground

	for _, pvc := range pvcs {
		// Ensure that our context is still active. It will be canceled if a
		// change to sts.Spec.Replicas is detected.
		select {
		case <-ctx.Done():
			return errors.New("concurrent statefulset modification detected")
		default:
		}

		p.Logger.Info("deleting PVC", "name", pvc.Name)
		if err := p.ClientSet.CoreV1().PersistentVolumeClaims(p.Namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
			// Wait for the underlying PV to be deleted before moving on to
			// the next volume.
			PropagationPolicy: &propagationPolicy,
			Preconditions: &metav1.Preconditions{
				// Ensure that this PVC is the same PVC that we slated for
				// deletion. If for some reason there are concurrent scale jobs
				// running, this will prevent us from re-deleting a PVC that
				// was removed and recreated.
				UID: &pvc.UID,
				// Ensure that this PVC has not changed since we fetched it.
				// This check doesn't help very much as a PVC is not actually
				// modified when it's mounted to a pod.
				ResourceVersion: &pvc.ResourceVersion,
			},
		}); err != nil {
			return errors.Wrapf(err, "delting pvc %s", pvc.Name)
		}
	}

	return nil
}
