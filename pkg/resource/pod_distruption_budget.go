package resource

import (
	"errors"

	"github.com/cockroachdb/cockroach-operator/pkg/labels"
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

	// TODO fix or should we use this?
	if pdb.ObjectMeta.Name == "" {
		pdb.ObjectMeta.Name = b.DiscoveryServiceName()
	}

	if pdb.ObjectMeta.Labels == nil {
		pdb.ObjectMeta.Labels = map[string]string{}
	}

	commonLabels := labels.Common(b.Cluster.cr)
	selector := commonLabels.Selector()
	pdb.Spec = policy.PodDisruptionBudgetSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: selector,
		},
	}

	// TODO should we always create a PDB?

	// Setup MinAvailable
	if b.Cluster.cr.Spec.MinAvailable != nil {
		minAvailable := b.Cluster.cr.Spec.MinAvailable
		minAvailableIS := intstr.FromInt(int(*minAvailable))
		pdb.Spec.MinAvailable = &minAvailableIS
	} else {
		// Setup MinAvailable as set or use the default value
		maxUnavailable := b.Cluster.cr.Spec.MaxUnavailable
		maxUnavailableIS := intstr.FromInt(int(*maxUnavailable))
		pdb.Spec.MaxUnavailable = &maxUnavailableIS
	}

	return nil
}

// TODO - what does this do???
func (b PdbBuilder) Placeholder() runtime.Object {
	return &policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.DiscoveryServiceName(),
		},
	}
}
