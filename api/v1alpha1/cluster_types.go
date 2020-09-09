/*
Copyright 2020 The Cockroach Authors

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

package v1alpha1

import (
	"github.com/operator-framework/operator-lib/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ConditionInstalling means the tools are installing on your cluster.
	ConditionInstalling status.ConditionType = "Installing"
	// ConditionComplete means the tools are installed on your cluster.
	ConditionComplete status.ConditionType = "Complete"
	// ConditionError means the installation has failed.
	ConditionError status.ConditionType = "Error"

	// Reasons for install
	ReasonStartInstall    status.ConditionReason = "StartInstall"
	ReasonInstallFinished status.ConditionReason = "FinishedInstall"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make" to regenerate code after modifying this file

// CrdblusterSpec defines the desired state of a CockroachDB Cluster
// that the operator maintains.
type CrdbClusterSpec struct {
	// Important: Run "make" to regenerate code after modifying this file
	// (Required) Number of nodes (pods) in the cluster
	Nodes int32 `json:"nodes"`
	// (Required) Container image information
	Image PodImage `json:"image"`
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
	// TODO document on how you need to set the max or the min
	// The maximum number of pods that can be unavailable during a rolling update.
	// This number is set in the PodDistruptionBudget and defaults to 1.
	MaxUnavailable *int32 `json:"maxUnavailable,omitempty"`
	// The min number of pods that can be unavailable during a rolling update.
	// This number is set in the PodDistruptionBudget and defaults to 1.
	MinAvailable *int32 `json:"minUnavailable,omitempty"`
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

	// List of internal operators conditions representing the current status of the cluster resource.
	// Currently, only `NotInitialized` is used
	OperatorConditions []ClusterCondition `json:"operatorConditions,omitempty"`

	// Database service version. Not populated and is just a placeholder currently.
	Version string `json:"version,omitempty"`

	// Conditions represent the latest available observations of an object's stateonfig
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes.conditions"
	// +optional
	Conditions status.Conditions `json:"conditions,omitempty"`
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
