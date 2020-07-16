package resource

import (
	"errors"
	"fmt"

	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PdbBuilder models the PodDistruptionBudget that the
// operator maintains.
type PdbBuilder struct {
	*Cluster

	Selector map[string]string
}

// Build creates a policy.PodDisruptionBudget for the
// StatefulSet.
func (b PdbBuilder) Build(obj runtime.Object) error {
	pdb, ok := obj.(*policy.PodDisruptionBudget)
	if !ok {
		return errors.New("failed to cast to PDB object")
	}

	// if we only have one Node we cannot have a PDB
	// TODO we need to validate this in the CRD API
	if b.Spec().Nodes == 1 {
		return nil
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

	// if we only have one Node we cannot have a PDB
	// TODO we need to validate this in the CRD API
	if b.Spec().Nodes == 1 {
		return nil
	}

	return &policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.DiscoveryServiceName(),
		},
	}
}
