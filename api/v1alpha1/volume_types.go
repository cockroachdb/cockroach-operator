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
	corev1 "k8s.io/api/core/v1"
)

// +kubebuilder:object:generate=true

// Volume defined storage configuration for the container with the Database.
// Only one of the fields should set
type Volume struct {
	// Directory from the host node's filesystem
	HostPath *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	// Temporary folder on the host node's filesystem
	EmptyDir *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// Persistent volume to use
	VolumeClaim *VolumeClaim `json:"pvc,omitempty"`
}

// +kubebuilder:object:generate=true
// VolumeClaim wraps a persistent volume claim (PVC) to use with the container.
// Only one of the fields should set
type VolumeClaim struct {
	// PVC to request a new persistent volume
	PersistentVolumeClaimSpec corev1.PersistentVolumeClaimSpec `json:"spec,omitempty"`
	// Existing PVC in the same namespace
	PersistentVolumeSource corev1.PersistentVolumeClaimVolumeSource `json:"source,omitempty"`
}
