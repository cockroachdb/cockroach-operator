package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CrdblusterSpec defines the desired state of Cluster
type CrdbClusterSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	// (Required) Number of nodes (pods) in the cluster
	Nodes int32 `json:"nodes"`
	// (Required) Container image with supported CockroachDB version
	Image string `json:"image,omitempty"`
	// (Optional) The database port (`--port` CLI parameter when starting the service)
	// Default: 26257
	GRPCPort *int32 `json:"grpcPort,omitempty"`
	// (Optional)  The web UI port (`--http-port` CLI parameter when starting the service)
	// Default: 8080
	HTTPPort *int32 `json:"httpPort,omitempty"`
	// (Optional) Whenever to start the cluster in secure or insecure mode. The secure mode
	// uses certificates from the secrets `NodeTLSSecret` and `ClientTLSSecret`.
	// If those are not specified, the operator generates certificates and
	// signs them using the cluster Certificate Authority. Such setup is not recommended
	// for production usage
	// Default: false
	TLSEnabled bool `json:"tlsEnabled,omitempty"`
	// (Optional) The secret with certificates and a private key for the TLS endpoint
	// on the database port. The standard naming of files is expected (tls.key, tls.crt, ca.crt)
	// Default: ""
	NodeTLSSecret string `json:"nodeTLSSecret,omitempty"`
	// (Optional) The secret with a certificate and a private key for root database user
	// Default: ""
	ClientTLSSecret string `json:"clientTLSSecret,omitempty"`
	// (Optional) The total size for caches (`--cache` command line parameter)
	// Default: "25%"
	Cache string `json:"cache,omitempty"`
	// (Optional) The maximum in-memory storage capacity available to store temporary
	// data for SQL queries (`--max-sql-memory` parameter)
	// Default: "25%"
	MaxSQLMemory string `json:"maxSQLMemory,omitempty"`
	// (Optional) Additional command line arguments for the `cockroach` binary
	// Default: ""
	AdditionalArgs []string `json:"additionalArgs,omitempty"`
	// (Optional) Database container resource limits. Any container limits
	// can be specified.
	// Default: (not specified)
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// (Required) Database disk storage configuration
	DataStore Volume `json:"dataStore,omitempty"`
}

// CrdbClusterStatus defines the observed state of Cluster
type CrdbClusterStatus struct {
	// Important: Run "make" to regenerate code after modifying this file

	// List of conditions representing the current status of the cluster resource.
	// Currently, only `NotInitialized` is used
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	// Database service version. Not populated and is just a placeholder currently.
	Version string `json:"version,omitempty"`
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
