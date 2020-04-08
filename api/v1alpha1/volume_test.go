package v1alpha1_test

import (
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestApply(t *testing.T) {
	metaMutator := func(string) metav1.ObjectMeta {
		return metav1.ObjectMeta{
			Name: "datadir",
		}
	}

	applyFn := func(vol *api.Volume, sts *appsv1.StatefulSetSpec) error {
		return vol.Apply("datadir", "test-container",
			"/data", sts, metaMutator)
	}

	errorsWith := func(message string) func(t *testing.T, vol *api.Volume, sts *appsv1.StatefulSetSpec) {
		return func(t *testing.T, vol *api.Volume, sts *appsv1.StatefulSetSpec) {
			err := applyFn(vol, sts)
			require.Error(t, err)
			assert.EqualError(t, err, message)
		}
	}

	sts := &appsv1.StatefulSetSpec{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "test-container",
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		sts      *appsv1.StatefulSetSpec
		vol      api.Volume
		assertFn func(t *testing.T, vol *api.Volume, sts *appsv1.StatefulSetSpec)
	}{
		{
			name: "both emptry dir and pvc claim provided",
			sts:  sts.DeepCopy(),
			vol: api.Volume{
				EmptyDir:    &corev1.EmptyDirVolumeSource{},
				VolumeClaim: &api.VolumeClaim{},
			},
			assertFn: errorsWith("one of HostPath, EmptyDir or VolumeClaim should be set"),
		},
		{
			name:     "no empty dir or pvc claim provided",
			sts:      sts.DeepCopy(),
			vol:      api.Volume{},
			assertFn: errorsWith("no valid Volume source provided"),
		},
		{
			name: "empty dir is correctly applied",
			sts:  sts.DeepCopy(),
			vol: api.Volume{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
			assertFn: func(t *testing.T, vol *api.Volume, sts *appsv1.StatefulSetSpec) {
				require.NoError(t, applyFn(vol, sts))

				assertVolumeMounts(t, sts, "datadir", "/data")

				require.Len(t, sts.Template.Spec.Volumes, 1)

				volume := &sts.Template.Spec.Volumes[0]
				require.NotNil(t, volume.EmptyDir)
				assert.Equal(t, "datadir", volume.Name)
			},
		},
		{
			name: "host dir is correctly applied",
			sts:  sts.DeepCopy(),
			vol: api.Volume{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/mnt/data",
				},
			},
			assertFn: func(t *testing.T, vol *api.Volume, sts *appsv1.StatefulSetSpec) {
				require.NoError(t, applyFn(vol, sts))
				assertVolumeMounts(t, sts, "datadir", "/data")

				require.Len(t, sts.Template.Spec.Volumes, 1)

				volume := &sts.Template.Spec.Volumes[0]
				require.NotNil(t, volume.HostPath)

				assert.Equal(t, "/mnt/data", volume.HostPath.Path)
				assert.Equal(t, "datadir", volume.Name)
			},
		},
		{
			name: "PVC is correctly applied",
			sts:  sts.DeepCopy(),
			vol: api.Volume{
				VolumeClaim: &api.VolumeClaim{
					PersistentVolumeClaimSpec: corev1.PersistentVolumeClaimSpec{},
					PersistentVolumeSource:    corev1.PersistentVolumeClaimVolumeSource{},
				},
			},
			assertFn: func(t *testing.T, vol *api.Volume, sts *appsv1.StatefulSetSpec) {
				require.NoError(t, applyFn(vol, sts))
				assertVolumeMounts(t, sts, "datadir", "/data")

				require.Len(t, sts.Template.Spec.Volumes, 1)

				volume := &sts.Template.Spec.Volumes[0]
				require.NotNil(t, volume.PersistentVolumeClaim)
				assert.Equal(t, "datadir", volume.Name)

				require.Len(t, sts.VolumeClaimTemplates, 1)

				claim := &sts.VolumeClaimTemplates[0]

				assert.Equal(t, corev1.PersistentVolumeClaimSpec{}, claim.Spec)
				assert.Equal(t, "datadir", claim.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertFn(t, &tt.vol, tt.sts)
		})
	}
}

func assertVolumeMounts(t *testing.T, sts *appsv1.StatefulSetSpec, name string, mount string) {
	require.Len(t, sts.Template.Spec.Containers[0].VolumeMounts, 1)

	volumeMnt := &sts.Template.Spec.Containers[0].VolumeMounts[0]
	assert.Equal(t, name, volumeMnt.Name)
	assert.Equal(t, mount, volumeMnt.MountPath)
}
