package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const NodeTLSSecretKeyword = "GENERATED"

// CrdblusterSpec defines the desired state of Cluster
type CrdbClusterSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	Nodes           int32  `json:"nodes"`
	Image           string `json:"image,omitempty"`
	GRPCPort        *int32 `json:"grpcPort,omitempty"`
	HTTPPort        *int32 `json:"httpPort,omitempty"`
	TLSEnabled      bool   `json:"tlsEnabled,omitempty"`
	NodeTLSSecret   string `json:"nodeTLSSecret,omitempty"`
	ClientTLSSecret string `json:"clientTLSSecret,omitempty"`
	// The total size for caches (--cache command line parameter)
	Cache string `json:"cache,omitempty"`
	// The maximum in-memory storage capacity available to store temporary
	// data for SQL queries (--max-sql-memory parameter)
	MaxSQLMemory string `json:"maxSQLMemory,omitempty"`
	// Optional command line args
	AdditionalArgs []string `json:"additionalArgs,omitempty"`
	// Resources set resource requests and limits for database containers
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	DataStore Volume                      `json:"dataStore,omitempty"`
	Topology  *Topology                   `json:"topology,omitempty"`
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
