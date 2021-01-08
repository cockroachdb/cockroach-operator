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

package v1alpha1

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestSetClusterSpecDefaults(t *testing.T) {
	original, ok := os.LookupEnv(RHEnvVar)
	// Ensure that RHEnvVar is reset after this test
	defer func() {
		if ok {
			os.Setenv(RHEnvVar, original)
		}
	}()

	maxUnavailable := int32(1)
	policy := v1.PullIfNotPresent

	testCases := []struct {
		Name  string
		Setup func()
		Error error
		In    CrdbClusterSpec
		Out   CrdbClusterSpec
	}{
		{
			Name: "Defaults",
			Setup: func() {
				os.Setenv(RHEnvVar, "RH_DEFAULT_IMAGE")
			},
			In: CrdbClusterSpec{},
			Out: CrdbClusterSpec{
				GRPCPort:       &DefaultGRPCPort,
				HTTPPort:       &DefaultHTTPPort,
				Cache:          "25%",
				MaxSQLMemory:   "25%",
				MaxUnavailable: &maxUnavailable,
				Image: PodImage{
					Name:           "RH_DEFAULT_IMAGE",
					PullPolicyName: &policy,
				},
			},
		},
		{
			Name: "NoOverrideImage",
			Setup: func() {
				os.Setenv(RHEnvVar, "RH_DEFAULT_IMAGE")
			},
			In: CrdbClusterSpec{Image: PodImage{Name: "Custom"}},
			Out: CrdbClusterSpec{
				GRPCPort:       &DefaultGRPCPort,
				HTTPPort:       &DefaultHTTPPort,
				Cache:          "25%",
				MaxSQLMemory:   "25%",
				MaxUnavailable: &maxUnavailable,
				Image: PodImage{
					Name:           "Custom",
					PullPolicyName: &policy,
				},
			},
		},
		{
			Name:  "ErrorOnNoImageNoEnvVar",
			Error: fmt.Errorf(".Image.Name and RELATED_IMAGE_COCKROACH are both unset"),
			Setup: func() {
				os.Unsetenv(RHEnvVar)
			},
			In: CrdbClusterSpec{},
			Out: CrdbClusterSpec{
				GRPCPort:       &DefaultGRPCPort,
				HTTPPort:       &DefaultHTTPPort,
				Cache:          "25%",
				MaxSQLMemory:   "25%",
				MaxUnavailable: &maxUnavailable,
				Image: PodImage{
					PullPolicyName: &policy,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup()

			err := SetClusterSpecDefaults(&tc.In)

			if tc.Error == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.Error.Error())
			}

			diff := cmp.Diff(tc.Out, tc.In)
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}
