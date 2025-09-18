package v1beta1

import (
	"time"

	"github.com/cockroachdb/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReconciliationMode describes the modus operandi of the CrdbCluster
// controller when reconciling CrdbNodes.
type ReconciliationMode string

// This file defines a CrdbCluster custom resource definition (CRD) object.
// NOTE: json tags are required.  Any new fields you add must have json tags for
// the fields to be serialized.
const (
	// CrdbClusterKind is the CrdbCluster CRD kind string.
	CrdbClusterKind = "CrdbCluster"

	// LocalPathStorageClass refers to the Rancher local-path-provisioner that
	// creates persistent volumes that utilize the local storage in each node.
	// See https://github.com/rancher/local-path-provisioner.
	// This is used for testing only.
	LocalPathStorageClass = "local-path"

	// MutableOnly ReconciliationMode reconciles mutable fields of all
	// CrdbNodes. New CrdbNodes will be created and CrdbNodes will be
	// decommissioned. MutableOnly is the default ReconciliationMode if one is
	// not specified.
	MutableOnly ReconciliationMode = "MutableOnly"

	// CreateOnly ReconciliationMode disables reconciliation for existing
	// CrdbNodes. New CrbdNodes nodes will be created inline with
	// CrdbNodeTemplate. CrdbNodes will not be decommissioned.
	CreateOnly ReconciliationMode = "CreateOnly"

	// Disabled ReconciliationMode disables reconciliation of CrdbNodes.
	// Changes to the CrdbNodeTemplate will not be propagated, new CrdbNodes
	// will not be created, and CrdbNodes will not be decommissioned.
	Disabled ReconciliationMode = "Disabled"
)

// CrdbNodeTemplate is the template from which CrdbNodes will be created or
// reconciled towards.
type CrdbNodeTemplate struct {
	// ObjectMeta is a set of metadata that will be propagated to CrdbNodes.
	// Only Labels and Annotations will be respected.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec is used as the template to construct or reconcile the Spec of
	// CrdbNodes, depending on the ReconciliationMode of the CrdbCluster.
	// +kubebuilder:validation:Required
	Spec CrdbNodeSpec `json:"spec,omitempty"`
}

// CrdbClusterSpec defines the desired state of CrdbCluster.
// NOTE: Run "make" to regenerate code after modifying this file.
type CrdbClusterSpec struct {
	// Mode sets the modus operandi of the CrdbCluster controller when
	// reconciling CrdbNodes.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=MutableOnly
	// +kubebuilder:validation:Enum=MutableOnly;CreateOnly;Disabled
	Mode *ReconciliationMode `json:"mode,omitempty"`

	// Template is the object that describes the CrdbNodes that will be created
	// and reconciled by the CrdbCluster controller, depending on the
	// ReconciliationMode.
	// +kubebuilder:validation:Optional
	Template CrdbNodeTemplate `json:"template"`

	// ClusterSettings is a set of CockroachDB CLUSTER SETTINGS that will be
	// set by the CrdbCluster controller via executing SET CLUSTER SETTING.
	// +kubebuilder:validation:Optional
	ClusterSettings map[string]string `json:"clusterSettings,omitempty"`

	// Regions specifies the regions in which this cluster is deployed, along
	// with information about how each region is configured.
	// +kubebuilder:validation:Required
	Regions []CrdbClusterRegion `json:"regions"`

	// TLSEnabled indicates whether the cluster is running in secure mode.
	// +kubebuilder:validation:Optional
	TLSEnabled bool `json:"tlsEnabled,omitempty"`

	// RollingRestartDelay is the delay between node restarts during a rolling
	// update. Defaults to 1 minute.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="1m"
	RollingRestartDelay *metav1.Duration `json:"rollingRestartDelay,omitempty"`
}

// Nodes returns the total number of CRDB nodes across all regions in the
// cluster.
func (s *CrdbClusterSpec) Nodes() int {
	cnt := 0
	for i := range s.Regions {
		cnt += int(s.Regions[i].Nodes)
	}
	return cnt
}

// Region returns the CrdbClusterRegion that matches the given name.
func (s *CrdbClusterSpec) Region(regionName string) *CrdbClusterRegion {
	for i := range s.Regions {
		if s.Regions[i].Code == regionName {
			return &s.Regions[i]
		}
	}
	return nil
}

// NodesInRegion returns the total number of CRDB nodes in a particular region
func (s *CrdbClusterSpec) NodesInRegion(regionName string) int32 {
	region := s.Region(regionName)
	if region != nil {
		return region.Nodes
	}
	return 0
}

// CrdbClusterRegion describes a region in which CRDB cluster nodes operate. It
// is used to generate the --join flag passed to each CrdbNode within the
// cluster.
type CrdbClusterRegion struct {
	// Code corresponds to the cloud provider's identifier of this region (e.g.
	// "us-east-1" for AWS, "us-east1" for GCP). This value is used to detect
	// which CrdbClusterRegion will be reconciled and must match the
	// "topology.kubernetes.io/region" label on Kubernetes Nodes in this
	// cluster.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Code string `json:"code"`

	// Nodes is the number of CRDB nodes that are in the region.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum:=0
	Nodes int32 `json:"nodes"`

	// CloudProvider sets the cloud provider for this region. When set, this value
	// is used to prefix the locality flag for all nodes in the region.
	// +kubebuilder:validation:Optional
	CloudProvider string `json:"cloudProvider,omitempty"`

	// Namespace is the name of the Kubernetes namespace that this
	// CrdbClusterRegion is deployed within. It is used to compute the --join
	// flag for this region. Defaults to the .Code of this region and then the
	// Namespace of this CrdbCluster, if not provided.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// Domain is the domain of the CrdbClusterRegion.
	// Other regions need to reach this region by connecting to
	// <cluster-name>.<namespace>.svc.<domain>.
	// It defaults an empty string, but this will not work
	// in a multi-region setup, where CrdbCluster objects are potentially
	// in different namespaces.
	// It will also not work if the k8s cluster has a custom domain.
	Domain string `json:"domain,omitempty"`
}

// CrdbClusterStatus defines the observed state of CrdbCluster.
// NOTE: Run "make" to regenerate code after modifying this file
type CrdbClusterStatus struct {
	// ObservedGeneration is the value of the ObjectMeta.Generation last
	// reconciled by the controller.
	// Note(alyshan): ObjectMeta.Generation uses int64, so we match the type.
	ObservedGeneration int64 `json:"observedGeneration"`

	// Actions are the set of operations taken on this cluster.
	Actions []ClusterAction `json:"actions,omitempty"`

	// Conditions are the set of current status indicators for the cluster.
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// Settings contains the cluster settings for the CRDB cluster.
	Settings map[string]string `json:"settings,omitempty"`

	// ReadyNodes is the number of nodes that are ready in this region.
	ReadyNodes int32 `json:"readyNodes,omitempty"`

	// Reconciled indicates whether the spec of ObservedGeneration is reconciled.
	Reconciled bool `json:"reconciled,omitempty"`

	// Provider is the name of the cloud provider that this object's k8s server is in.
	Provider string `json:"provider,omitempty"`

	// Region is the name of the region that this crdbcluster object's k8s server is in.
	// This is useful for consumers to determine if this region's crdb pods
	// are ready, etc..
	Region string `json:"region,omitempty"`

	// Image is the CockroachDB image currently running in this cluster.
	// +kubebuilder:validation:Optional
	Image string `json:"image,omitempty"`

	// Version is the version of CockroachDB currently running in this cluster.
	// This is populated by specifing the version where version is the output of executing
	// `cockroach version` command on running pods.
	// +kubebuilder:validation:Optional
	Version string `json:"version,omitempty"`
}

// SetAction adds a new action to the cluster actions list if no matching
// action exists, or else overwrites the existing action.
func (s *CrdbClusterStatus) SetAction(typ ActionType, st ActionStatus) {
	action := ClusterAction{
		Type:               typ,
		Status:             st,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	for i := range s.Actions {
		if s.Actions[i].Type == typ {
			s.Actions[i] = action
			return
		}
	}
	s.Actions = append(s.Actions, action)
}

// HasAction returns true if an action of the given type exists in the
// action list and has the given status value.
func (s *CrdbClusterStatus) HasAction(typ ActionType, st ActionStatus) bool {
	for i := range s.Actions {
		if s.Actions[i].Type == typ {
			return s.Actions[i].Status == st
		}
	}
	return false
}

// ActionPresent returns true if an action of the given type exists in the
// action list.
func (s *CrdbClusterStatus) ActionPresent(typ ActionType) bool {
	for i := range s.Actions {
		if s.Actions[i].Type == typ {
			return true
		}
	}
	return false
}

// GetCondition returns the condition status, last transition time
// of the given type, and "true" if the given condition type exists.
// Otherwise, it returns empty values for the former two and "false".
func (s *CrdbClusterStatus) GetCondition(
	typ ClusterConditionType,
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
func (s *CrdbClusterStatus) HasCondition(typ ClusterConditionType, st metav1.ConditionStatus) bool {
	cond, _, ok := s.GetCondition(typ)
	return ok && cond == st
}

// SetCondition adds a new condition to the cluster condition list if no matching
// condition exists yet, or else overwrites the existing condition.
func (s *CrdbClusterStatus) SetCondition(typ ClusterConditionType, status metav1.ConditionStatus) {
	cond := ClusterCondition{
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

// ClusterCondition describes the current state of some aspect of the cluster's
// status. The operator will add these to the Conditions list as it completes
// work.
type ClusterCondition struct {
	// Type is the kind of this condition.
	// +kubebuilder:validation:Required
	Type ClusterConditionType `json:"type"`
	// Status is the current state of the condition: True, False or Unknown.
	// +kubebuilder:validation:Required
	Status metav1.ConditionStatus `json:"status"`
	// LastTransitionTime is the time at which the condition was last updated.
	// +kubebuilder:validation:Required
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

// ClusterAction describes an operation performed by the operator on the
// CrdbCluster.
type ClusterAction struct {
	// Type is the kind of this action.
	// +kubebuilder:validation:Required
	Type ActionType `json:"type"`
	// Status is the current state of the action: Starting, Failed, Finished or Unknown.
	// +kubebuilder:validation:Required
	Status ActionStatus `json:"status"`
	// LastTransitionTime is the time at which the condition was last updated.
	// +kubebuilder:validation:Required
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:webhook:conversion
// CrdbCluster is the Schema for the crdbclusters API.
type CrdbCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CrdbClusterSpec   `json:"spec,omitempty"`
	Status CrdbClusterStatus `json:"status,omitempty"`
}

// IsInitialized returns true when the cluster has been marked as initialized.
func (cluster *CrdbCluster) IsInitialized() bool {
	return cluster.Status.HasCondition(ClusterInitialized, metav1.ConditionTrue)
}

// IsReady returns nil once all nodes in the CRDB cluster are ready
// to serve traffic.
func (cluster *CrdbCluster) IsReady() error {
	if cluster.Generation > cluster.Status.ObservedGeneration {
		return errors.Newf("observed CrdbCluster generation %d is lower "+
			"than expected generation %d", cluster.Status.ObservedGeneration, cluster.Generation)
	}
	if !cluster.Status.HasCondition(ClusterInitialized, metav1.ConditionTrue) {
		return errors.Newf("cluster is not initialized")
	}
	return nil
}

// +kubebuilder:object:root=true

// CrdbClusterList contains a list of CrdbCluster.
type CrdbClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CrdbCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CrdbCluster{}, &CrdbClusterList{})
}