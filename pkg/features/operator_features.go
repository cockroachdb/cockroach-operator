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
	// PartitionedUpdate controls how the statefulset is updated
	PartitionedUpdate featuregate.Feature = "PartitionedUpdate"

	// owner: @alina
	// alpha: v0.1
	// beta: v1.0
	// Decommission controls how the statefulset scales down
	Decommission featuregate.Feature = "UseDecommission"

	// alpha: v0.1
	// beta: v1.0
	// Upgrades controls upgrade mechanism
	Upgrade featuregate.Feature = "Upgrade"

	// owner: @chrislovecnm
	// alpha: v0.1
	// beta: v1.0
	// ResizePVC allows for the resizing of PVC Volumes
	ResizePVC featuregate.Feature = "ResizePVC"

	// alpha: v0.1
	// beta: v1.0
	// CrdbVersionValidator controls upgrade mechanism
	CrdbVersionValidator featuregate.Feature = "CrdbVersionValidator"

	// owner: @chrislovecnm
	// alpha: v0.1
	// beta: v1.0
	// GA: v1.7.7
	// GenerateCerts uses crdb binary to generate self signed certifcates
	GenerateCerts featuregate.Feature = "GenerateCerts"
	// owner: @alina
	// alpha: v0.1
	// beta: v1.0
	// RestartCluster uses crdb binary to generate self signed certifcates
	ClusterRestart featuregate.Feature = "ClusterRestart"
	// owner: @alina
	// alpha: v0.1
	// beta: v1.0
	// PVCPruner uses crdb binary to prune PVC on decommission
	PVCPruner featuregate.Feature = "PVCPruner"
)

func init() {
	runtime.Must(utilfeature.DefaultMutableFeatureGate.Add(defaultOperatorFeatureGates))
}

// defaultKubernetesFeatureGates consists of all known operator-specific feature keys.
// To add a new feature, define a key for it above and add it here. The features will be
// available throughout Kubernetes binaries.
var defaultOperatorFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	PartitionedUpdate: {Default: true, PreRelease: featuregate.Beta},
	// Turing of Decomission because of design bug at this point
	Decommission:         {Default: false, PreRelease: featuregate.Alpha},
	ResizePVC:            {Default: true, PreRelease: featuregate.Beta},
	CrdbVersionValidator: {Default: true, PreRelease: featuregate.Beta},
	GenerateCerts:        {Default: true, PreRelease: featuregate.GA},
	ClusterRestart:       {Default: true, PreRelease: featuregate.Beta},
	PVCPruner:            {Default: false, PreRelease: featuregate.Beta},
}
