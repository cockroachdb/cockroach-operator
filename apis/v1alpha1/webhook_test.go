/*
Copyright 2025 The Cockroach Authors

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

package v1alpha1_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestCrdbClusterDefault(t *testing.T) {
	cluster := &CrdbCluster{
		Spec: CrdbClusterSpec{
			Image: &PodImage{},
		},
	}

	maxUnavailable := int32(1)
	policy := v1.PullIfNotPresent
	expected := CrdbClusterSpec{
		GRPCPort:       &DefaultGRPCPort,
		HTTPPort:       &DefaultHTTPPort,
		SQLPort:        &DefaultSQLPort,
		MaxUnavailable: &maxUnavailable,
		Image:          &PodImage{PullPolicyName: &policy},
	}

	ctx := context.Background()
	_ = cluster.Default(ctx, cluster)
	require.Equal(t, expected, cluster.Spec)
}

func TestValidateIngress(t *testing.T) {

	tests := []struct {
		name     string
		cluster  *CrdbCluster
		expected []error
	}{
		{
			name: "ingress config with UI host missing",
			cluster: &CrdbCluster{
				Spec: CrdbClusterSpec{
					Ingress: &IngressConfig{UI: &Ingress{IngressClassName: "abc"}},
				},
			},
			expected: []error{fmt.Errorf("host required for UI")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			require.Equal(t, tt.expected, tt.cluster.ValidateIngress())
		})

	}
}

func TestCreateCrdbCluster(t *testing.T) {
	block := v1.PersistentVolumeBlock
	fs := v1.PersistentVolumeFilesystem
	testcases := []struct {
		Cluster *CrdbCluster
		ErrMsg  string
	}{
		{
			Cluster: &CrdbCluster{
				Spec: CrdbClusterSpec{
					Image: &PodImage{},
				},
			},
			ErrMsg: "you have to provide the cockroachDBVersion or cockroach image",
		},
		{
			Cluster: &CrdbCluster{
				Spec: CrdbClusterSpec{
					Image:              &PodImage{Name: "testImage"},
					CockroachDBVersion: "v2.1.20",
				},
			},
			ErrMsg: "you have provided both cockroachDBVersion and cockroach image, please provide only one",
		},
		{
			Cluster: &CrdbCluster{Spec: CrdbClusterSpec{
				Image: &PodImage{Name: "testImage"},
				DataStore: Volume{
					VolumeClaim: &VolumeClaim{
						PersistentVolumeClaimSpec: v1.PersistentVolumeClaimSpec{},
					},
				},
			}},
			ErrMsg: "you have not provided pvc.volumeMode value.",
		},
		{
			Cluster: &CrdbCluster{Spec: CrdbClusterSpec{
				Image: &PodImage{Name: "testImage"},
				DataStore: Volume{
					VolumeClaim: &VolumeClaim{
						PersistentVolumeClaimSpec: v1.PersistentVolumeClaimSpec{
							VolumeMode: &block,
						},
					},
				},
			}},
			ErrMsg: "you have provided unsupported pvc.volumeMode, currently only Filesystem is supported.",
		},
		{
			Cluster: &CrdbCluster{Spec: CrdbClusterSpec{
				Image: &PodImage{Name: "testImage"},
				DataStore: Volume{
					VolumeClaim: &VolumeClaim{
						PersistentVolumeClaimSpec: v1.PersistentVolumeClaimSpec{
							VolumeMode: &fs,
						},
					},
				},
			}},
			ErrMsg: "",
		},
	}

	ctx := context.Background()
	for _, testcase := range testcases {
		_, err := testcase.Cluster.ValidateCreate(ctx, testcase.Cluster)
		if testcase.ErrMsg == "" {
			require.NoError(t, err)
			continue
		}
		require.Error(t, err)
		require.Equal(t, err.Error(), testcase.ErrMsg)
	}
}

func TestUpdateCrdbCluster(t *testing.T) {
	oldCluster := CrdbCluster{
		Spec: CrdbClusterSpec{
			Image: &PodImage{},
		},
	}

	testcases := []struct {
		Cluster *CrdbCluster
		ErrMsg  string
	}{
		{
			Cluster: &CrdbCluster{
				Spec: CrdbClusterSpec{},
			},
			ErrMsg: "you have to provide the cockroachDBVersion or cockroach image",
		},
		{
			Cluster: &CrdbCluster{
				Spec: CrdbClusterSpec{
					Image:              &PodImage{Name: "testImage"},
					CockroachDBVersion: "v2.1.20",
				},
			},
			ErrMsg: "you have provided both cockroachDBVersion and cockroach image, please provide only one",
		},
	}

	ctx := context.Background()
	for _, testcase := range testcases {
		_, err := testcase.Cluster.ValidateUpdate(ctx, &oldCluster, testcase.Cluster)
		require.Error(t, err)
		require.Equal(t, err.Error(), testcase.ErrMsg)
	}
}

func TestUpdateCrdbClusterLabels(t *testing.T) {
	oldCluster := CrdbCluster{
		Spec: CrdbClusterSpec{
			Image: &PodImage{},
			AdditionalLabels: map[string]string{
				"k": "v",
			},
		},
	}
	fs := v1.PersistentVolumeFilesystem

	testcases := []struct {
		Cluster     *CrdbCluster
		ShouldError bool
	}{
		{
			Cluster: &CrdbCluster{Spec: CrdbClusterSpec{
				Image:            &PodImage{Name: "testImage"},
				AdditionalLabels: map[string]string{"k": "v"},
				DataStore: Volume{
					VolumeClaim: &VolumeClaim{
						PersistentVolumeClaimSpec: v1.PersistentVolumeClaimSpec{
							VolumeMode: &fs,
						},
					},
				},
			}},
			ShouldError: false,
		},
		{
			Cluster: &CrdbCluster{Spec: CrdbClusterSpec{
				Image:            &PodImage{Name: "testImage"},
				AdditionalLabels: map[string]string{"k": "x"},
				DataStore: Volume{
					VolumeClaim: &VolumeClaim{
						PersistentVolumeClaimSpec: v1.PersistentVolumeClaimSpec{
							VolumeMode: &fs,
						},
					},
				},
			}},
			// label k has a different value.
			ShouldError: true,
		},
		{
			Cluster: &CrdbCluster{Spec: CrdbClusterSpec{
				Image: &PodImage{Name: "testImage"},
				DataStore: Volume{
					VolumeClaim: &VolumeClaim{
						PersistentVolumeClaimSpec: v1.PersistentVolumeClaimSpec{
							VolumeMode: &fs,
						},
					},
				},
			}},
			// labels are missing / empty.
			ShouldError: true,
		},
		{
			Cluster: &CrdbCluster{Spec: CrdbClusterSpec{
				Image:            &PodImage{Name: "testImage"},
				AdditionalLabels: map[string]string{"k": "v", "kk": "v"},
				DataStore: Volume{
					VolumeClaim: &VolumeClaim{
						PersistentVolumeClaimSpec: v1.PersistentVolumeClaimSpec{
							VolumeMode: &fs,
						},
					},
				},
			}},
			// labels contain additional kv.
			ShouldError: true,
		},
	}

	ctx := context.Background()
	for _, tc := range testcases {
		_, err := tc.Cluster.ValidateUpdate(ctx, &oldCluster, tc.Cluster)
		if tc.ShouldError {
			require.Error(t, err)
			require.Equal(t, err.Error(), "mutating additionalLabels field is not supported")
		} else {
			require.NoError(t, err)
		}
	}
}
