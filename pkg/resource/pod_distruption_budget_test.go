/*
Copyright 2021 The Cockroach Authors

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
	"testing"

	"fmt"

	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPDBBuilder(t *testing.T) {
	var maxUnavailable int32 = 3
	cluster := testutil.NewBuilder("test-cluster").Namespaced("test-ns").WithMaxUnavailable(&maxUnavailable)
	commonLabels := labels.Common(cluster.Cr())
	selector := commonLabels.Selector(cluster.Cr())

	maxUnavailableIS := intstr.FromInt(3)

	tests := []struct {
		name     string
		cluster  *resource.Cluster
		selector map[string]string
		expected *policy.PodDisruptionBudget
	}{
		{
			name:     "builds default discovery service",
			cluster:  cluster.Cluster(),
			selector: commonLabels.Selector(cluster.Cr()),
			expected: &policy.PodDisruptionBudget{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-cluster",
					Labels: map[string]string{},
				},
				Spec: policy.PodDisruptionBudgetSpec{
					MaxUnavailable: &maxUnavailableIS,
					Selector: &metav1.LabelSelector{
						MatchLabels: selector,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := &policy.PodDisruptionBudget{}

			err := resource.PdbBuilder{
				Cluster:  tt.cluster,
				Selector: tt.selector,
			}.Build(actual)
			require.NoError(t, err)

			diff := cmp.Diff(tt.expected, actual, testutil.RuntimeObjCmpOpts...)
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}
