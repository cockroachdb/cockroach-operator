/*
Copyright 2023 The Cockroach Authors

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

package actor

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//  backoffFactory is a replacable global for backoff creation. It may be
// replaced with shorter times to allow testing of Wait___ functions without
// waiting the entire default period
var backoffFactory = defaultBackoffFactory

func defaultBackoffFactory(maxTime time.Duration) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxTime
	return b
}

// newResizePVC creates and returns a new resizePVC struct
func newResizePVC(scheme *runtime.Scheme, cl client.Client, clientset kubernetes.Interface) Actor {
	return &resizePVC{
		action: newAction(scheme, cl, nil, clientset),
	}
}

// resizePVC resizes a PVC
type resizePVC struct {
	action
}

//GetActionType returns api.RequestCertAction action used to set the cluster status errors
func (rp *resizePVC) GetActionType() api.ActionType {
	return api.ResizePVCAction
}

// Act in this implementation resizes PVC volumes of a CR sts.
func (rp *resizePVC) Act(ctx context.Context, cluster *resource.Cluster, log logr.Logger) error {
	// If we do not have a volume claim we do not have PVCs
	if cluster.Spec().DataStore.VolumeClaim == nil {
		log.Info("Skipping PVC resize as VolumeClaim does not exist")
		return nil
	}

	// Get the sts and compare the sts size to the size in the CR
	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      cluster.StatefulSetName(),
	}
	statefulSet := &appsv1.StatefulSet{}
	if err := rp.client.Get(ctx, key, statefulSet); err != nil {
		return errors.Wrap(err, "failed to fetch statefulset")
	}

	// TODO statefulSetIsUpdating is not quite working as expected.
	// I had to check status.  We should look at the update code in partition update to address this
	if statefulSetIsUpdating(statefulSet) {
		return NotReadyErr{Err: errors.New("resize statefulset is updating, waiting for the update to finish")}
	}

	status := &statefulSet.Status
	if status.CurrentReplicas == 0 || status.CurrentReplicas < status.Replicas {
		log.Info("resize pvc statefulset does not have all replicas up")
		return NotReadyErr{Err: errors.New("resize pvc statefulset does not have all replicas up")}
	}

	// Maybe this should be an error since we should not have this, but I wanted to check anyways
	if len(statefulSet.Spec.VolumeClaimTemplates) == 0 {
		log.Info("Skipping PVC resize as PVCs do not exist")
		return nil
	}

	stsStorageSizeDeployed := statefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage()
	stsStorageSizeSet := cluster.Spec().DataStore.VolumeClaim.PersistentVolumeClaimSpec.Resources.Requests.Storage()

	// If the sizes match do not resize
	if stsStorageSizeDeployed.Equal(stsStorageSizeSet.DeepCopy()) {
		log.Info("Skipping PVC resize as sizes match")
		return nil
	}

	log.Info("Starting PVC resize")

	// Find all of the PVCs and resize them
	if err := rp.findAndResizePVC(ctx, statefulSet, cluster, rp.clientset, log); err != nil {
		return errors.Wrapf(err, "updating PVCs for statefulset %s.%s", cluster.Namespace(), cluster.StatefulSetName())
	}

	log.Info("Starting updating sts")

	// Update the STS with the correct volume size, in case more pods are created
	// We will create a copy and update the copy, and then delete the original without
	// deleting the Pods.  The new sts is then used to create a new statefulset.
	if err := rp.updateSts(ctx, statefulSet, cluster, log); err != nil {
		return errors.Wrapf(err, "updating statefulset %s.%s", cluster.Namespace(), cluster.StatefulSetName())
	}

	// TODO this is not working so we will need to patch the sts
	// with a value that will force a restart
	// We are thinking that patching an annotation will help
	/*
		if !cluster.Spec().DataStore.SupportsAutoResize {
			log.Info("Starting rolling sts")
			// Roll the entire STS in order for the Pods to resize
			if err := rp.rollSts(ctx, cluster, clientset); err != nil {
				return errors.Wrapf(err, "error restarting statefulset %s.%s", cluster.Namespace(), cluster.StatefulSetName())
			}
		} else {
			log.Info("Volumes support autoresizing so not restarting STS Pods")
		}*/

	log.Info("PVC resize completed")
	return nil
}

// updateSts updates the size of an STS' VolumeClaimTemplate to match the new size in the CR.
// In order to update the volume claim template we have to delete the STS without cascading and then
// create the sts.
func (rp *resizePVC) updateSts(ctx context.Context, sts *appsv1.StatefulSet, cluster *resource.Cluster, log logr.Logger) error {

	// delete the original sts, but do not delete the Pods
	orphan := metav1.DeletePropagationOrphan
	if err := rp.client.Delete(ctx, sts, &client.DeleteOptions{PropagationPolicy: &orphan}); err != nil {
		return err
	}

	f := func() error {
		return rp.recreateSTS(ctx, cluster, log)
	}

	b := backoffFactory(5 * time.Minute)
	return backoff.Retry(f, backoff.WithContext(b, ctx))
}

func (rp *resizePVC) recreateSTS(ctx context.Context, cluster *resource.Cluster, log logr.Logger) error {
	// Use same StatefulSetBuilder that we run in Deploy to
	// rebuild and save the StatefulSet with the new PVC size
	r := resource.NewManagedKubeResource(ctx, rp.client, cluster, kube.AnnotatingPersister)
	_, err := (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.StatefulSetBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(cluster.Spec().AdditionalLabels),
		},
		Owner:  cluster.Unwrap(),
		Scheme: rp.scheme,
	}).Reconcile()

	if err != nil {
		log.Info("unable to re-create sts, will retry")
	}
	return err
}

// findAndResizePVC finds all active PVCs and resizes them to the new size contained in the cluster
// definition.
func (rp *resizePVC) findAndResizePVC(ctx context.Context, sts *appsv1.StatefulSet, cluster *resource.Cluster,
	clientset kubernetes.Interface, log logr.Logger) error {
	// K8s doesn't provide a way to tell if a PVC or PV is currently in use by
	// a pod. However, it is safe to assume that any PVCs with an ordinal great
	// than or equal to the sts' Replicas is not in use. As only pods with with
	// an ordinal < Replicas will exist. Any PVCs with an ordinal less than
	// Replicas is in use. To detect this, we build a map of PVCs that we
	// consider to be in use and skip and PVCs that it contains
	// the name of.
	log.Info("starting finding and resizing all PVCs")
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
		return errors.Wrap(err, "converting statefulset selector to metav1 selector")
	}

	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(cluster.Namespace()).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return errors.Wrap(err, "finding PVCs to for resizing")
	}

	log.Info("resizing PVCs")
	for _, pvc := range pvcs.Items {
		// Resize PVCs that are still in use
		if pvcsToKeep[pvc.Name] {
			size := cluster.Spec().DataStore.VolumeClaim.PersistentVolumeClaimSpec.Resources.Requests.Storage()
			pvc.Spec.Resources.Requests[v1.ResourceStorage] = *size

			if _, err := clientset.CoreV1().PersistentVolumeClaims(cluster.Namespace()).Update(ctx, &pvc, metav1.UpdateOptions{}); err != nil {
				return errors.Wrap(err, "error resizing PVCs")
			}

			log.Info(fmt.Sprintf("resized %s", pvc.Name))
		}
	}

	log.Info("found and resized all PVCs")
	return nil
}
