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

import v1 "k8s.io/api/core/v1"

var (
	DefaultGRPCPort       int32 = 26257
	DefaultHTTPPort       int32 = 8080
	DefaultMaxUnavailable int32 = 1
)

// SetClusterSpecDefaults sets various values to the
// default vars set above.
func SetClusterSpecDefaults(cs *CrdbClusterSpec) {
	if cs.GRPCPort == nil {
		cs.GRPCPort = &DefaultGRPCPort
	}

	if cs.HTTPPort == nil {
		cs.HTTPPort = &DefaultHTTPPort
	}

	if cs.MaxUnavailable == nil && cs.MinAvailable == nil {
		cs.MaxUnavailable = &DefaultMaxUnavailable
	}

	if cs.Cache == "" {
		cs.Cache = "25%"
	}

	if cs.MaxSQLMemory == "" {
		cs.MaxSQLMemory = "25%"
	}

	if cs.Image.PullPolicyName == nil {
		policy := v1.PullIfNotPresent
		cs.Image.PullPolicyName = &policy
	}

}
