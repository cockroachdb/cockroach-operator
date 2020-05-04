package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

// +kubebuilder:object:generate=true

type Volume struct {
	HostPath    *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	EmptyDir    *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	VolumeClaim *VolumeClaim                 `json:"pvc,omitempty"`
}

// +kubebuilder:object:generate=true

type VolumeClaim struct {
	PersistentVolumeClaimSpec corev1.PersistentVolumeClaimSpec         `json:"spec,omitempty"`
	PersistentVolumeSource    corev1.PersistentVolumeClaimVolumeSource `json:"source,omitempty"`
}
