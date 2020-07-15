package resource

import (
	"errors"
	"fmt"

	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// This service only exists to create DNS entries for each pod in
// the StatefulSet such that they can resolve each other's IP addresses.
// It does not create a load-balanced ClusterIP and should not be used directly
// by clients in most circumstances.
type PdbBuilder struct {
	*Cluster

	Selector map[string]string
}

func (b PdbBuilder) Build(obj runtime.Object) error {
	pdb, ok := obj.(*policy.PodDisruptionBudget)
	if !ok {
		return errors.New("failed to cast to PDB object")
	}

	// TODO fix or should we use this?
	if pdb.ObjectMeta.Name == "" {
		pdb.ObjectMeta.Name = b.DiscoveryServiceName()
	}

	if pdb.ObjectMeta.Labels == nil {
		pdb.ObjectMeta.Labels = map[string]string{}
	}

	// TODO test sevice name
	// Can we use the common label selector?
	// commonLabels := labels.Common(cluster.Cr())
	//  selector: commonLabels.Selector(),

	labelSelector, err := metav1.ParseToLabelSelector("app=" + b.DiscoveryServiceName())
	if err != nil {
		return fmt.Errorf("unexpected error parsing label: %v", err)
	}

	// TODO check to see if both are not nil??

	if b.Cluster.cr.Spec.MinAvailable != nil {
		min := b.Cluster.cr.Spec.MinAvailable
		minIS := intstr.FromInt(int(*min))
		pdb.Spec = policy.PodDisruptionBudgetSpec{
			MinAvailable: &minIS,
			Selector:     labelSelector,
		}
	} else if b.Cluster.cr.Spec.MaxUnavailable != nil {
		max := b.Cluster.cr.Spec.MaxUnavailable
		maxIS := intstr.FromInt(int(*max))
		pdb.Spec = policy.PodDisruptionBudgetSpec{
			Selector:       labelSelector,
			MaxUnavailable: &maxIS,
		}
	} else {
		maxIS := intstr.FromInt(1)
		pdb.Spec = policy.PodDisruptionBudgetSpec{
			Selector:       labelSelector,
			MaxUnavailable: &maxIS,
		}
	}

	return nil
}

func (b PdbBuilder) Placeholder() runtime.Object {
	return &policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.DiscoveryServiceName(),
		},
	}
}
