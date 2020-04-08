package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

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

// CrdblusterSpec defines the desired state of Cluster
type CrdbClusterSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	Nodes     int32     `json:"nodes"`
	Image     string    `json:"image,omitempty"`
	GRPCPort  *int32    `json:"grpcPort,omitempty"`
	HTTPPort  *int32    `json:"httpPort,omitempty"`
	DataStore Volume    `json:"dataStore,omitempty"`
	Topology  *Topology `json:"topology,omitempty"`
}

// CrdbClusterStatus defines the observed state of Cluster
type CrdbClusterStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	Version    string             `json:"version,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=all;cockroachdb
// +kubebuilder:subresource:status

// CrdbCluster is the Schema for the clusters API
type CrdbCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CrdbClusterSpec   `json:"spec,omitempty"`
	Status CrdbClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=all;addons

// CrdbClusterList contains a list of Cluster
type CrdbClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CrdbCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CrdbCluster{}, &CrdbClusterList{})
}
