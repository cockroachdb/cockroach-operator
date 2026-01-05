/*
Copyright 2026 The Cockroach Authors

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

package update

import (
	"context"
	"reflect"
	"testing"

	semver "github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestIsPatch(t *testing.T) {
	tests := []struct {
		description    string
		wantVersion    *semver.Version
		currentVersion *semver.Version
		result         bool
	}{
		{
			"isPatch true",
			semver.MustParse("20.1.6"),
			semver.MustParse("20.1.5"),
			true,
		},
		{
			"isPatch false",
			semver.MustParse("19.2.6"),
			semver.MustParse("20.1.5"),
			false,
		},
		{
			"isPatch false test two",
			semver.MustParse("19.2.6"),
			semver.MustParse("19.3.5"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, isPatch(test.wantVersion, test.currentVersion) == test.result, "patch result failed")
		})
	}

}

func TestIsMajorUpgradeAllowed(t *testing.T) {
	tests := []struct {
		description    string
		wantVersion    *semver.Version
		currentVersion *semver.Version
		result         bool
	}{
		{
			"forward minor update",
			semver.MustParse("20.1.6"),
			semver.MustParse("20.1.5"),
			false,
		},
		{
			"forward major update",
			semver.MustParse("20.1.6"),
			semver.MustParse("19.2.5"),
			true,
		},
		{
			"backward minor version",
			semver.MustParse("19.2.6"),
			semver.MustParse("19.3.5"),
			false,
		},
		{
			"is backward major version",
			semver.MustParse("19.2.6"),
			semver.MustParse("21.3.5"),
			false,
		},
		{
			"is two major",
			semver.MustParse("22.2.6"),
			semver.MustParse("19.3.5"),
			false,
		},
		{
			"is skipping innovative release same year",
			semver.MustParse("24.3"),
			semver.MustParse("24.1"),
			true,
		},
		{
			"is skipping innovative release next year",
			semver.MustParse("25.2"),
			semver.MustParse("24.3"),
			true,
		},
		{
			"is skipping innovative and regular release",
			semver.MustParse("25.1"),
			semver.MustParse("24.1"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, isMajorUpgradeAllowed(test.wantVersion, test.currentVersion) == test.result, "isMajorUpgradeAllowed test failed")
		})
	}
}

func TestIsMajorRollbackAllowed(t *testing.T) {
	tests := []struct {
		description    string
		wantVersion    *semver.Version
		currentVersion *semver.Version
		result         bool
	}{
		{
			"forward minor update",
			semver.MustParse("20.1.6"),
			semver.MustParse("20.1.5"),
			false,
		},
		{
			"backward major version 20 to 19",
			semver.MustParse("19.2.6"),
			semver.MustParse("20.1.5"),
			true,
		},
		{
			"backward major version 21 to 20",
			semver.MustParse("20.2.6"),
			semver.MustParse("21.1.5"),
			true,
		},
		{
			"is two major",
			semver.MustParse("22.2.6"),
			semver.MustParse("19.3.5"),
			false,
		},
		{
			"is skipping innovative release for same year",
			semver.MustParse("24.1"),
			semver.MustParse("24.3"),
			true,
		},
		{
			"is skipping innovative release previous year",
			semver.MustParse("24.3"),
			semver.MustParse("25.2"),
			true,
		},
		{
			"is skipping innovative and regular release",
			semver.MustParse("24.1"),
			semver.MustParse("25.1"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, isMajorRollbackAllowed(test.wantVersion, test.currentVersion) == test.result, "isMajorRollbackAllowed test failed")
		})
	}
}

func TestGetNextReleases(t *testing.T) {
	tests := []struct {
		description    string
		currentVersion string
		result         []string
	}{
		{
			"returns the possible upgrade targets for 24.1",
			"24.1",
			[]string{"24.2", "24.3"},
		},
		{
			"returns the possible upgrade targets for 24.3",
			"24.3",
			[]string{"25.1", "25.2"},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, reflect.DeepEqual(getNextReleases(test.currentVersion), test.result),
				"getNextReleases test failed")
		})
	}
}

func TestGetPreviousReleases(t *testing.T) {
	tests := []struct {
		description    string
		currentVersion string
		result         []string
	}{
		{
			"returns the possible rollback targets for 24.3",
			"24.3",
			[]string{"24.2", "24.1"},
		},
		{
			"returns the possible upgrade targets for 25.2",
			"25.2",
			[]string{"25.1", "24.3"},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, reflect.DeepEqual(getPreviousReleases(test.currentVersion), test.result),
				"getPreviousReleases test failed")
		})
	}
}

func TestGenerateReleases(t *testing.T) {
	tests := []struct {
		description string
		uptoYear    int
		result      []string
	}{
		{
			"returns possible releases till 2026",
			25,
			[]string{"24.1", "24.2", "24.3", "25.1", "25.2", "25.3", "25.4"},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, reflect.DeepEqual(generateReleases(test.uptoYear), test.result), "generateReleases test failed")
		})
	}
}

func TestGetReleaseType(t *testing.T) {
	tests := []struct {
		description  string
		major, minor int
		result       ReleaseType
	}{
		{
			"returns releaseType of a 24.2",
			24, 2,
			Innovative,
		},
		{
			"returns releaseType of a 25.1",
			25, 1,
			Innovative,
		},
		{
			"returns releaseType of a 24.3",
			24, 3,
			Regular,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, reflect.DeepEqual(getReleaseType(test.major, test.minor), test.result),
				"getReleaseType test failed")
		})
	}
}

func TestMakeUpdateCockroachVersionFunction_FIPSImageHandling(t *testing.T) {
	tests := []struct {
		name           string
		cockroachImage string
		version        string
		expectedImage  string
		description    string
	}{
		{
			name:           "SHA256_image_preserved",
			cockroachImage: "registry.connect.redhat.com/cockroachdb/cockroach@sha256:3b78a4d3864e16d270979d9d7756012abf09e6cdb7f1eb4832f6497c26281099",
			version:        "v25.2.1",
			expectedImage:  "registry.connect.redhat.com/cockroachdb/cockroach@sha256:3b78a4d3864e16d270979d9d7756012abf09e6cdb7f1eb4832f6497c26281099",
			description:    "FIPS SHA256 image should be preserved exactly as-is",
		},
		{
			name:           "Regular_Docker_Hub_image",
			cockroachImage: "cockroachdb/cockroach:v25.2.1",
			version:        "v25.2.1",
			expectedImage:  "cockroachdb/cockroach:v25.2.1",
			description:    "Regular Docker Hub images should work normally",
		},
		{
			name:           "Custom_registry_with_tag",
			cockroachImage: "my-registry.com/cockroachdb/cockroach:v25.2.1",
			version:        "v25.2.1",
			expectedImage:  "my-registry.com/cockroachdb/cockroach:v25.2.1",
			description:    "Custom registry images with tags should be preserved",
		},
		{
			name:           "Fips_Image_Preserved",
			cockroachImage: "cockroachdb/cockroach:v25.2.1-fips",
			version:        "v25.2.1",
			expectedImage:  "cockroachdb/cockroach:v25.2.1-fips",
			description:    "FIPS images should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a StatefulSet with both main and init containers
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-cluster",
					Namespace:   "test-namespace",
					Annotations: make(map[string]string),
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Name:  "db-init",
									Image: "old-image:v25.1.0",
								},
							},
							Containers: []corev1.Container{
								{
									Name:  "db",
									Image: "old-image:v25.1.0",
								},
							},
						},
					},
				},
			}

			// Create the update function
			updateFunc := makeUpdateCockroachVersionFunction(tt.cockroachImage, tt.version, "v25.1.0")

			// Apply the update
			updatedSts, err := updateFunc(sts)
			require.NoError(t, err, tt.description)
			require.NotNil(t, updatedSts)

			// Verify annotations are set correctly
			assert.Equal(t, tt.version, updatedSts.Annotations[resource.CrdbVersionAnnotation],
				"Version annotation should be set correctly")
			assert.Equal(t, tt.expectedImage, updatedSts.Annotations[resource.CrdbContainerImageAnnotation],
				"Container image annotation should preserve the exact image reference")

			// Verify main container image is updated
			assert.Equal(t, tt.expectedImage, updatedSts.Spec.Template.Spec.Containers[0].Image,
				"Main container image should be updated to the exact image reference")

			// Verify init container image is updated
			assert.Equal(t, tt.expectedImage, updatedSts.Spec.Template.Spec.InitContainers[0].Image,
				"Init container image should be updated to the exact image reference")
		})
	}
}

func TestMakeUpdateCockroachVersionFunction_NoInitContainers(t *testing.T) {
	// Test case where StatefulSet has no init containers (TLS disabled)
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-cluster",
			Namespace:   "test-namespace",
			Annotations: make(map[string]string),
		},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					// No InitContainers
					Containers: []corev1.Container{
						{
							Name:  "db",
							Image: "cockroachdb/cockroach:v25.1.0-fips",
						},
					},
				},
			},
		},
	}

	fipsImage := "cockroachdb/cockroach:v25.2.1-fips"
	updateFunc := makeUpdateCockroachVersionFunction(fipsImage, "v25.2.1", "v25.1.0")

	updatedSts, err := updateFunc(sts)
	require.NoError(t, err)
	require.NotNil(t, updatedSts)

	// Should not crash and should update main container
	assert.Equal(t, fipsImage, updatedSts.Spec.Template.Spec.Containers[0].Image)
	assert.Len(t, updatedSts.Spec.Template.Spec.InitContainers, 0, "Should have no init containers")
}

func TestMakeIsCRBPodIsRunningNewVersionFunction_WithInitContainers(t *testing.T) {
	tests := []struct {
		name               string
		cockroachImage     string
		mainContainerImage string
		initContainerImage string
		podReady           bool
		expectError        bool
		description        string
	}{
		{
			name:               "All_containers_updated_and_ready",
			cockroachImage:     "cockroachdb/cockroach:v25.2.1",
			mainContainerImage: "cockroachdb/cockroach:v25.2.1",
			initContainerImage: "cockroachdb/cockroach:v25.2.1",
			podReady:           true,
			expectError:        false,
			description:        "When all containers are updated and pod is ready, should succeed",
		},
		{
			name:               "Main_container_not_updated",
			cockroachImage:     "cockroachdb/cockroach:v25.2.1",
			mainContainerImage: "cockroachdb/cockroach:v25.2.0",
			initContainerImage: "cockroachdb/cockroach:v25.2.1",
			podReady:           true,
			expectError:        true,
			description:        "When main container is not updated, should return error",
		},
		{
			name:               "Init_container_not_updated",
			cockroachImage:     "cockroachdb/cockroach:v25.2.1",
			mainContainerImage: "cockroachdb/cockroach:v25.2.0",
			initContainerImage: "cockroachdb/cockroach:v25.2.0",
			podReady:           true,
			expectError:        true,
		},
		{
			name:               "Pod_not_ready",
			cockroachImage:     "cockroachdb/cockroach:v25.2.1",
			mainContainerImage: "cockroachdb/cockroach:v25.2.1",
			initContainerImage: "cockroachdb/cockroach:v25.2.1",
			podReady:           false,
			expectError:        true,
			description:        "When pod is not ready, should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pod with both main and init containers
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sts-0",
					Namespace: "test-namespace",
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:  "db-init",
							Image: tt.initContainerImage,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "db",
							Image: tt.mainContainerImage,
						},
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{},
				},
			}

			// Set pod ready condition based on test case
			if tt.podReady {
				pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				})
			}

			// Create fake clientset and add the pod
			clientset := fake.NewSimpleClientset(pod)

			// Create StatefulSet
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sts",
					Namespace: "test-namespace",
				},
			}

			// Create UpdateSts struct
			updateSts := &UpdateSts{
				ctx:       context.Background(),
				clientset: clientset,
				sts:       sts,
			}

			// Create verification function
			verifyFunc := makeIsCRDBPodIsRunningNewVersionFunction(tt.cockroachImage)

			// Test the verification
			err := verifyFunc(updateSts, 0, logf.Log)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestMakeIsCRBPodIsRunningNewVersionFunction_NoInitContainers(t *testing.T) {
	tests := []struct {
		name               string
		cockroachImage     string
		mainContainerImage string
		podReady           bool
		expectError        bool
		description        string
	}{
		{
			name:               "All_containers_updated_and_ready",
			cockroachImage:     "cockroachdb/cockroach:v25.2.1",
			mainContainerImage: "cockroachdb/cockroach:v25.2.1",
			podReady:           true,
			expectError:        false,
			description:        "When all containers are updated and pod is ready, should succeed",
		},
		{
			name:               "Main_container_not_updated",
			cockroachImage:     "cockroachdb/cockroach:v25.2.1",
			mainContainerImage: "cockroachdb/cockroach:v25.2.0",
			podReady:           true,
			expectError:        true,
			description:        "When main container is not updated, should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test pod without init containers (TLS disabled)
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sts-0",
					Namespace: "test-namespace",
				},
				Spec: corev1.PodSpec{
					// No InitContainers
					Containers: []corev1.Container{
						{
							Name:  "db",
							Image: tt.mainContainerImage,
						},
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{},
				},
			}

			if tt.podReady {
				pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				})
			}

			// Create fake clientset and add the pod
			clientset := fake.NewSimpleClientset(pod)

			// Create StatefulSet
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sts",
					Namespace: "test-namespace",
				},
			}

			// Create UpdateSts struct
			updateSts := &UpdateSts{
				ctx:       context.Background(),
				clientset: clientset,
				sts:       sts,
			}

			verifyFunc := makeIsCRDBPodIsRunningNewVersionFunction(tt.cockroachImage)

			// Should succeed - no init containers to check
			err := verifyFunc(updateSts, 0, logf.Log)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}
