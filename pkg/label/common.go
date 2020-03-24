package label

import (
	"fmt"
	"github.com/cockroachlabs/crdb-operator/api/v1alpha1"
)

// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
const (
	// The name of a higher level application this one is part of
	NameLabelKey = "app.kubernetes.io/name"
	// A unique name identifying the instance of an application
	InstanceLabelKey = "app.kubernetes.io/instance"
	// The current version of the application
	VersionLabelKey = "app.kubernetes.io/version"
	// The component within the architecture
	ComponentLabelKey = "app.kubernetes.io/component"
	// The name of a higher level application this one is part of
	PartOfLabelKey = "app.kubernetes.io/part-of"
	// The tool being used to manage the operation of an application
	ManagedByLabelKey = "app.kubernetes.io/managed-by"
)

func MakeCommonLabels(c *v1alpha1.CrdbCluster) map[string]string {
	labels := c.Labels

	// keep part-of customized if it was set by high-level app
	if _, ok := labels[PartOfLabelKey]; !ok {
		labels[PartOfLabelKey] = "cockroachdb"
	}

	labels[NameLabelKey] = "cockroachdb"
	labels[InstanceLabelKey] = fmt.Sprintf("%s/%s", c.Namespace, c.Name)
	labels[VersionLabelKey] = c.Status.Version
	labels[ComponentLabelKey] = "database"
	labels[ManagedByLabelKey] = "crdb-operator"

	return labels
}
