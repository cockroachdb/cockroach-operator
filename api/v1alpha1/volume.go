package v1alpha1

import (
	"fmt"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (v *Volume) Apply(name string, container string, path string,
	spec *appsv1.StatefulSetSpec, metaMutator func(name string) metav1.ObjectMeta) error {
	sourcesNum := sourcesSet(v)
	if sourcesNum > 1 {
		return errors.New("one of HostPath, EmptyDir or VolumeClaim should be set")
	}

	if sourcesNum == 0 {
		return errors.New("no valid Volume source provided")
	}

	if err := v.applyToPod(name, container, path, &spec.Template.Spec); err != nil {
		return err
	}

	if v.VolumeClaim != nil {
		if spec.VolumeClaimTemplates == nil {
			spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{}
		}

		pvc := corev1.PersistentVolumeClaim{
			ObjectMeta: metaMutator(v.VolumeClaim.PersistentVolumeSource.ClaimName),
			Spec:       v.VolumeClaim.PersistentVolumeClaimSpec,
			//Status: corev1.PersistentVolumeClaimStatus{`
			//	Phase: corev1.ClaimPending,
			//},
		}

		spec.VolumeClaimTemplates = append(spec.VolumeClaimTemplates, pvc)
	}

	return nil
}

func (v *Volume) applyToPod(name string, container string, path string, spec *corev1.PodSpec) error {
	found := false
	for i, _ := range spec.Containers {
		c := &spec.Containers[i]
		if c.Name == container {
			found = true
			if c.VolumeMounts == nil {
				c.VolumeMounts = []corev1.VolumeMount{}
			}

			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      name,
				MountPath: path,
			})
			break
		}
	}

	if !found {
		return fmt.Errorf("failed to find container %s to attach volume", container)
	}

	volume := corev1.Volume{
		Name: name,
	}

	if v.HostPath != nil {
		volume.VolumeSource = corev1.VolumeSource{
			HostPath: v.HostPath,
		}
	} else if v.EmptyDir != nil {
		volume.VolumeSource = corev1.VolumeSource{
			EmptyDir: v.EmptyDir,
		}
	} else if v.VolumeClaim != nil {
		volume.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &v.VolumeClaim.PersistentVolumeSource,
		}
	}

	if spec.Volumes == nil {
		spec.Volumes = []corev1.Volume{}
	}

	spec.Volumes = append(spec.Volumes, volume)

	return nil
}

func sourcesSet(v *Volume) int {
	set := 0

	if v.HostPath != nil {
		set += 1
	}

	if v.EmptyDir != nil {
		set += 1
	}

	if v.VolumeClaim != nil {
		set += 1
	}

	return set
}
