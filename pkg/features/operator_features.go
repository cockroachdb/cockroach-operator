/*
Copyright 2020 The Cockroach Authors

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

package features

import (
	featuregate "github.com/cockroachdb/cockroach-operator/pkg/featuregates"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"k8s.io/apimachinery/pkg/util/runtime"
)

const (
	// Every feature gate should add method here following this template:
	//
	// // owner: @username
	// // alpha: v1.4
	// MyFeature() bool

	// owner: @chrislovecnm
	// alpha: v0.1
	// beta: v1.0
	//
	// RunAsNonRoot controls whether the operator should create a statefulset as another user
	RunAsNonRoot featuregate.Feature = "RunAsNonRoot"
)

func init() {
	runtime.Must(utilfeature.DefaultMutableFeatureGate.Add(defaultKubernetesFeatureGates))
}

// defaultKubernetesFeatureGates consists of all known Kubernetes-specific feature keys.
// To add a new feature, define a key for it above and add it here. The features will be
// available throughout Kubernetes binaries.
var defaultKubernetesFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	RunAsNonRoot: {Default: true, PreRelease: featuregate.Alpha},
}
