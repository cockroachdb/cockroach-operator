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

package v1alpha1_test

import (
	"fmt"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestCrdbClusterDefault(t *testing.T) {
	cluster := &CrdbCluster{}

	maxUnavailable := int32(1)
	policy := v1.PullIfNotPresent
	expected := CrdbClusterSpec{
		GRPCPort:       &DefaultGRPCPort,
		HTTPPort:       &DefaultHTTPPort,
		SQLPort:        &DefaultSQLPort,
		MaxUnavailable: &maxUnavailable,
		Image:          PodImage{PullPolicyName: &policy},
	}

	cluster.Default()
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
