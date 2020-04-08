package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterConditionType string

const (
	InitializedCondition ClusterConditionType = "Initialized"
)

// TODO(vladdy): this should eventually be converged with
//     https://github.com/kubernetes/enhancements/pull/1624

type ClusterCondition struct {
	// +required
	Type ClusterConditionType `json:"type"`
	// +required
	Status metav1.ConditionStatus `json:"status"`
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}
