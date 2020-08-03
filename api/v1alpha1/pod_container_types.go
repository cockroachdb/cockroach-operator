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

import v1 "k8s.io/api/core/v1"

// +kubebuilder:object:generate=true

// PodImage represents the image information for a container that is used
// to build the StatefulSet.
type PodImage struct {
	// (Required) Container image with supported CockroachDB version
	Name string `json:"name,omitempty"`
	// PullPolicy for the image
	PullPolicyName *v1.PullPolicy `json:"pullPolicy,omitempty"`
	// Secret name containing the dockerconfig to use for a
	// registry that requires authentication. The secret
	// must be configured first by the user.
	PullSecret *string `json:"pullSecret,omitempty"`
}
