package v1beta1

import (
	"time"

	"github.com/cockroachdb/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeConditionType string

const (
	// CrdbNodeKind is the CrdbNode CRD kind string.
	CrdbNodeKind = "CrdbNode"

	// PodReady is a duplicate of Kubernetes' Pod's Ready Condition but
	// reported from the CrdbNode controller. If true, it implies that
	// PodRunning is also true.
	PodReady NodeConditionType = "PodReady"

	// PodReady is a duplicate of Kubernetes' Pod's Running Condition but
	// reported from the CrdbNode controller.
	PodRunning NodeConditionType = "PodRunning"

	// Default ports.
	DefaultCockroachGRPCPort = 26258
	DefaultCockroachHTTPPort = 8080
	DefaultCockroachSQLPort  = 26257
)

// CrdbNodeSpec defines the desired state of an individual
// CockroachDB node process.
type CrdbNodeSpec struct {
	// Image is the location and name of the CockroachDB container image to
	// deploy (e.g. cockroachdb/cockroach:v20.1.1).
	// +kubebuilder:validation:Required
	Image string `json:"image,omitempty"`

	// DataStore specifies the disk configuration for this CrdbNode.
	// +kubebuilder:validation:Required
	DataStore DataStore `json:"dataStore,omitempty"`

	// Certificates controls the TLS configuration of this CrdbNode.
	// +kubebuilder:validation:Required
	Certificates Certificates `json:"certificates,omitempty"`

	// ResourceRequirements is the resource requirements for the main
	// crdb container.
	// +kubebuilder:validation:Required
	ResourceRequirements corev1.ResourceRequirements `json:"resourceRequirements,omitempty"`

	// ServiceAccountName is the name of a service account to use for
	// running pods.
	// +kubebuilder:validation:Required
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Flags specify the flags that will be used for starting the cluster.
	// +kubebuilder:validation:Optional
	Flags map[string]string `json:"flags,omitempty"`

	// GRPCPort is the port used for the grpc service. This controls the
	// `--listen-addr` flag to `cockroach start`. If not set, default to 26258.
	// +kubebuilder:validation:Optional
	GRPCPort *int32 `json:"grpcPort,omitempty"`

	// HTTPPort is the port used for the http service. This controls the
	// `--http-addr` flag to `cockroach start`. If not set, default to 8080.
	// +kubebuilder:validation:Optional
	HTTPPort *int32 `json:"httpPort,omitempty"`

	// SQLPort is the port used for the sql service. This controls the
	// `--sql-addr` flag to `cockroach start`. If not set, default to 26257.
	// +kubebuilder:validation:Optional
	SQLPort *int32 `json:"sqlPort,omitempty"`

	// PodLabels are the labels that should be applied to the underlying CRDB
	// pod.
	// +kubebuilder:validation:Optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// PodAnnotations are the annotations that should be applied to the
	// underlying CRDB pod.
	// +kubebuilder:validation:Optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// Env is a list of environment variables to set in the container.
	// +kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Tolerations is the set of tolerations to apply to a node.
	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// NodeSelector is the set of nodeSelector labels to apply to a node.
	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity is the affinity policy to apply to a node.
	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// TopologySpreadConstraints will be used to construct the crdb pod's topology spread
	// constraints.
	// +kubebuilder:validation:Optional
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
}

// DataStore is the specification for the disk configuration of a CrdbNode.
// Exactly one of VolumeSource or VolumeClaimTemplate must be specified.
type DataStore struct {
	// VolumeSource specifies a pre-existing VolumeSource to be mounted to this
	// CrdbNode.
	// +kubebuilder:validation:Optional
	VolumeSource *corev1.VolumeSource `json:"dataStore,omitempty"`
	// VolumeClaimTemplate specifies a that a new PersistentVolumeClaim should
	// be created for this CrdbNode.
	// +kubebuilder:validation:Optional
	VolumeClaimTemplate *corev1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// Certificates controls the TLS configuration of the various services run by
// CockroachDB. Only one of its members may be specified.
type Certificates struct {
	// ExternalCertificates allows end users to assume full control over the
	// certificates used by CockroachDB.
	ExternalCertificates *ExternalCertificates `json:"externalCertificates,omitempty"`
}

// ExternalCertificates is a certificate configuration that mounts the
// pre-existing and non-operator controlled Kubernetes Secrets into CrdbNode
// Pods for use as TLS on the various services run by CockroachDB.
type ExternalCertificates struct {
	// CAConfigMapName is the name of a Kubernetes ConfigMap containing a ca.crt
	// entry that was used to sign other external certificates. This is used to
	// validate the node and client certificates.
	// +kubebuilder:validation:Optional
	CAConfigMapName string `json:"caConfigMapName,omitempty"`

	// HTTPSecretName is the name of a Kubernetes TLS Secret that will be used
	// for the HTTP service.
	// +kubebuilder:validation:Optional
	HTTPSecretName string `json:"httpSecretName,omitempty"`

	// NodeSecretName is the name of a Kubernetes TLS Secret that will be used
	// when receiving incoming connections from other nodes for RPC and SQL calls.
	// +kubebuilder:validation:Required
	NodeSecretName string `json:"nodeSecretName,omitempty"`

	// RootSQLClientSecretName is the name of a Kubernetes TLS secret holding SQL
	// client certificates for the root SQL user. It allows the operator to perform
	// various administrative actions (e.g. set cluster settings).
	// +kubebuilder:validation:Required
	RootSQLClientSecretName string `json:"rootSqlClientSecretName,omitempty"`
}

func (extCerts *ExternalCertificates) Validate() error {
	if extCerts.CAConfigMapName == "" {
		return errors.New("caConfigMapName must be specified")
	}
	return nil
}

// GetGRPCPort returns the port used to serve grpc connections, defaulting to
// 26258 if not specified.
func (s *CrdbNodeSpec) GetGRPCPort() int32 {
	if s.GRPCPort != nil {
		return *s.GRPCPort
	}
	return DefaultCockroachGRPCPort
}

// GetHTTPPort returns the port used to serve http connections, defaulting to
// 8080 if not specified.
func (s *CrdbNodeSpec) GetHTTPPort() int32 {
	if s.HTTPPort != nil {
		return *s.HTTPPort
	}
	return DefaultCockroachHTTPPort
}

// GetSQLPort returns the port used to serve grpc connections, defaulting to
// 26257 if not specified.
func (s *CrdbNodeSpec) GetSQLPort() int32 {
	if s.SQLPort != nil {
		return *s.SQLPort
	}
	return DefaultCockroachSQLPort
}

// CrdbNodeStatus defines the observed state of an individual
// CockroachDB node process.
type CrdbNodeStatus struct {
	// ObservedGeneration is the value of the ObjectMeta.Generation last
	// reconciled by the controller.
	ObservedGeneration int64 `json:"observedGeneration"`

	// Conditions are the set of current status indicators for the node.
	Conditions []NodeCondition `json:"conditions,omitempty"`
}

// GetCondition returns the condition status, last transition time
// of the given type, and "true" if the given condition type exists.
// Otherwise, it returns empty values for the former two and "false".
func (s *CrdbNodeStatus) GetCondition(
	typ NodeConditionType,
) (st metav1.ConditionStatus, ts time.Time, exists bool) {
	for i := range s.Conditions {
		if s.Conditions[i].Type == typ {
			return s.Conditions[i].Status, s.Conditions[i].LastTransitionTime.Time, true
		}
	}
	exists = false
	return
}

// HasCondition returns true if a condition of the given type exists in the
// condition list and has the given status value.
func (s *CrdbNodeStatus) HasCondition(typ NodeConditionType, st metav1.ConditionStatus) bool {
	cond, _, ok := s.GetCondition(typ)
	return ok && cond == st
}

// SetCondition adds a new condition to the cluster condition list if no matching
// condition exists yet, or else overwrites the existing condition.
func (s *CrdbNodeStatus) SetCondition(typ NodeConditionType, status metav1.ConditionStatus) {
	cond := NodeCondition{
		Type:               typ,
		Status:             status,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	for i := range s.Conditions {
		if s.Conditions[i].Type == typ {
			s.Conditions[i] = cond
			return
		}
	}
	s.Conditions = append(s.Conditions, cond)
}

type NodeCondition struct {
	// Type is the kind of this condition.
	// +kubebuilder:validation:Required
	Type NodeConditionType `json:"type"`
	// Status is the current state of the condition: True, False or Unknown.
	// +kubebuilder:validation:Required
	Status metav1.ConditionStatus `json:"status"`
	// LastTransitionTime is the time at which the condition was last updated.
	// +kubebuilder:validation:Required
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CrdbNode is the Schema for the crdbnode API.
type CrdbNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CrdbNodeSpec   `json:"spec,omitempty"`
	Status CrdbNodeStatus `json:"status,omitempty"`
}

// IsReady returns nil once the underlying Pod running CockroachDB is ready.
func (node *CrdbNode) IsReady() error {
	if node.Generation > node.Status.ObservedGeneration {
		return errors.Newf(
			"observed %T generation %d is lower than expected generation %d",
			node,
			node.Status.ObservedGeneration,
			node.Generation,
		)
	}

	if !node.Status.HasCondition(PodRunning, metav1.ConditionTrue) {
		return errors.Newf(
			"%T's pod is not running yet",
			node,
		)
	}
	if !node.Status.HasCondition(PodReady, metav1.ConditionTrue) {
		return errors.Newf(
			"%T's pod is not ready yet",
			node,
		)
	}
	return nil
}

// +kubebuilder:object:root=true

// CrdbNodeList contains a list of CrdbCluster.
type CrdbNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CrdbNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CrdbNode{}, &CrdbNodeList{})
}