package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterConditionType string

const (
	NotInitializedCondition ClusterConditionType = "NotInitialized"
)

// TODO(vladdy): this should eventually be converged with
//     https://github.com/kubernetes/enhancements/pull/1624

// ClusterCondition represents cluster status as it is perceived by
// the operator
type ClusterCondition struct {
	// +required
	// Type/Name of the condition
	Type ClusterConditionType `json:"type"`
	// +required
	// Condition status: True, False or Unknown
	Status metav1.ConditionStatus `json:"status"`
	// +required
	// The time when the condition was updated
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}
