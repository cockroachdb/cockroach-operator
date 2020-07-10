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
