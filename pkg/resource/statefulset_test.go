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

package resource_test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/ptr"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var update = flag.Bool("update", false, "update the golden files of this test")

const migrationLabel = "crdb.io/migrating"

func TestStatefulSetBuilder(t *testing.T) {
	// turn on featuregate to test rules
	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("AffinityRules=true,TolerationRules=true"))
	sc := testutil.InitScheme(t)

	decoder, encoder := testutil.Yamlizers(t, sc)

	inputSuffix := "_in.yaml"
	folder := filepath.Join("testdata", filepath.FromSlash(t.Name()))
	testInputs := filepath.Join(folder, "/*"+inputSuffix)
	clusterFiles, err := filepath.Glob(testInputs)
	if err != nil || len(clusterFiles) == 0 {
		t.Fatalf("failed to find cluster specs %s: %v", testInputs, err)
	}

	for _, inFile := range clusterFiles {
		testName := inFile[len(folder)+1 : len(inFile)-len(inputSuffix)]
		clusterObj := decoder(load(t, inFile))
		cr, ok := clusterObj.(*api.CrdbCluster)
		if !ok {
			t.Fatal("failed to deserialize CrdbCluster")
		}
		commonLabels := labels.Common(cr)
		cluster := resource.NewCluster(cr)

		t.Run(testName, func(t *testing.T) {
			actual := &appsv1.StatefulSet{}

			err := resource.StatefulSetBuilder{
				Cluster:   &cluster,
				Selector:  commonLabels.Selector(cluster.Spec().AdditionalLabels),
				Telemetry: "kubernetes-operator-gke",
			}.Build(actual)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = encoder(actual, &buf)
			require.NoError(t, err)

			expectedStr := testutil.ReadOrUpdateGoldenFile(t, buf.String(), *update)
			expected := decoder([]byte(expectedStr))

			diff := cmp.Diff(expected, actual, testutil.RuntimeObjCmpOpts...)
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}

func TestRHImage(t *testing.T) {
	rhImage := "redhat-coachroach-test:v22"
	// os.Setenv(resource.RhEnvVar, rhImage)

	cluster := resource.NewCluster(&api.CrdbCluster{
		Spec: api.CrdbClusterSpec{
			Image: &api.PodImage{},
		},
	})

	b := resource.StatefulSetBuilder{
		Cluster: &cluster,
	}

	container := b.MakeContainers()
	//we should run the version checker to set this field so it is empty
	if container[0].Image != "" {
		assert.Fail(t, fmt.Sprintf("unexpected result expected image to equal: %s, got: %s", rhImage, container[0].Image))
	}

}

func load(t *testing.T, file string) []byte {
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to load yaml file %s: %v", file, err)
	}

	return content
}

func TestStatefulSetStopScaleUpWithMigrationLabel(t *testing.T) {
	tests := []struct {
		name             string
		crNodes          int32
		currentReplicas  int32
		labels           map[string]string
		expectedReplicas int32
	}{
		{
			name:             "Normal Scale Up - No Label",
			crNodes:          3,
			currentReplicas:  2,
			labels:           nil,
			expectedReplicas: 3,
		},
		{
			name:            "Stop Scale Up - Label Present",
			crNodes:         3,
			currentReplicas: 2,
			labels: map[string]string{
				migrationLabel: "true",
			},
			expectedReplicas: 2,
		},
		{
			name:            "Allow Scale Down - Label Present",
			crNodes:         3,
			currentReplicas: 4,
			labels: map[string]string{
				migrationLabel: "true",
			},
			expectedReplicas: 3,
		},
		{
			name:            "Stop Scale Up - False Label Value",
			crNodes:         3,
			currentReplicas: 2,
			labels: map[string]string{
				migrationLabel: "false",
			},
			expectedReplicas: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &api.CrdbCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
					Labels:    tt.labels,
				},
				Spec: api.CrdbClusterSpec{
					Nodes:              tt.crNodes,
					CockroachDBVersion: "v20.2.5",
					Image: &api.PodImage{
						Name: "cockroachdb/cockroach:v20.2.5",
					},
					DataStore: api.Volume{
						VolumeClaim: &api.VolumeClaim{
							PersistentVolumeClaimSpec: corev1.PersistentVolumeClaimSpec{},
						},
					},
				},
			}

			cluster := resource.NewCluster(cr)
			commonLabels := labels.Common(cr)

			// Pre-populate the existing StatefulSet with current replicas
			actual := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.Int32(tt.currentReplicas),
				},
			}

			builder := resource.StatefulSetBuilder{
				Cluster:   &cluster,
				Selector:  commonLabels.Selector(cluster.Spec().AdditionalLabels),
				Telemetry: "kubernetes-operator-test",
			}

			err := builder.Build(actual)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedReplicas, *actual.Spec.Replicas)
		})
	}
}
