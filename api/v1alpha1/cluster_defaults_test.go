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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestSetClusterSpecDefaults(t *testing.T) {
	s := &CrdbClusterSpec{}
	maxUnavailable := int32(1)
	policy := v1.PullIfNotPresent
	expected := &CrdbClusterSpec{
		GRPCPort:       &DefaultGRPCPort,
		HTTPPort:       &DefaultHTTPPort,
		Cache:          "25%",
		MaxSQLMemory:   "25%",
		MaxUnavailable: &maxUnavailable,
		Image: PodImage{
			PullPolicyName: &policy,
		},
	}

	SetClusterSpecDefaults(s)

	diff := cmp.Diff(expected, s)
	if diff != "" {
		assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
	}
}
