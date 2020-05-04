package v1alpha1

// AvailabilityZone represents one failure domain within the cluster
type AvailabilityZone struct {
	// Locality to use
	Locality string `json:"locality"`
	// Suffix to add to add to stateful set name
	StatefulSetSuffix string `json:"suffix,omitempty"`
	//Labels to target Kubernetes nodes
	Labels map[string]string `json:"labels,omitempty"`
}

type Topology struct {
	// List of availability zones
	Zones []AvailabilityZone `json:"zones,omitempty"`
}
