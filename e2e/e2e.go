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

package e2e

import (
	"time"
)

var (
	// CreateClusterTimeout is the amount of time to wait for a new cluster to be
	// ready.
	CreateClusterTimeout = 20 * time.Minute
)

// Some common values used in e2e test suites.
const (
	MinorVersion1        = "cockroachdb/cockroach:v24.1.0"
	MinorVersion2        = "cockroachdb/cockroach:v24.1.2"
	MajorVersion         = "cockroachdb/cockroach:v24.2.2"
	NonExistentVersion   = "cockroachdb/cockroach-non-existent:v21.1.999"
	SkipFeatureVersion   = "cockroachdb/cockroach:v20.1.0"
	InvalidImage         = "nginx:latest"
	DefaultCPULimit      = "800m"
	DefaultMemoryLimit   = "3Gi"
	DefaultCPURequest    = "500m"
	DefaultMemoryRequest = "1Gi"
)
