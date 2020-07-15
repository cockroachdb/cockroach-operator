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
