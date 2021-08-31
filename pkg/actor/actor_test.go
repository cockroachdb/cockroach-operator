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

package actor_test

import (
	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"testing"
)

func containsAction(actions []api.ActionType, action api.ActionType) bool {
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}

func TestDecommissionFeatureGate(t *testing.T) {
	// Setup fake client
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()
	director := actor.ClusterDirector{}

	cluster.SetTrue(api.InitializedCondition)

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true")
	actions := director.GetActionsToExecute(cluster)
	require.True(t, containsAction(actions, api.DecommissionAction))

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=false")
	actions = director.GetActionsToExecute(cluster)
	require.False(t, containsAction(actions, api.DecommissionAction))
}

func TestVersionValidatorFeatureGate(t *testing.T) {
	// Setup fake client
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()
	director := actor.ClusterDirector{}

	cluster.SetTrue(api.InitializedCondition)

	utilfeature.DefaultMutableFeatureGate.Set("CrdbVersionValidator=true")
	actions := director.GetActionsToExecute(cluster)
	require.True(t, containsAction(actions, api.VersionCheckerAction))

	utilfeature.DefaultMutableFeatureGate.Set("CrdbVersionValidator=false")
	actions = director.GetActionsToExecute(cluster)
	require.False(t, containsAction(actions, api.VersionCheckerAction))
}

func TestResizePVCFeatureGate(t *testing.T) {
	// Setup fake client
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()
	director := actor.ClusterDirector{}

	cluster.SetTrue(api.InitializedCondition)

	utilfeature.DefaultMutableFeatureGate.Set("ResizePVC=true")
	actions := director.GetActionsToExecute(cluster)
	require.True(t, containsAction(actions, api.ResizePVCAction))

	utilfeature.DefaultMutableFeatureGate.Set("ResizePVC=false")
	actions = director.GetActionsToExecute(cluster)
	require.False(t, containsAction(actions, api.ResizePVCAction))
}

func TestClusterRestartFeatureGate(t *testing.T) {
	// Setup fake client
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()
	director := actor.ClusterDirector{}

	cluster.SetTrue(api.InitializedCondition)
	cluster.SetTrue(api.CrdbVersionChecked)

	utilfeature.DefaultMutableFeatureGate.Set("ClusterRestart=true")
	actions := director.GetActionsToExecute(cluster)
	require.True(t, containsAction(actions, api.ClusterRestartAction))

	utilfeature.DefaultMutableFeatureGate.Set("ClusterRestart=false")
	actions = director.GetActionsToExecute(cluster)
	require.False(t, containsAction(actions, api.ClusterRestartAction))
}

func TestTotallyUninitialized(t *testing.T) {
	// Setup fake client
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()
	director := actor.ClusterDirector{}

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")

	actions := director.GetActionsToExecute(cluster)
	require.True(t, cmp.Equal(actions, []api.ActionType{api.VersionCheckerAction, api.GenerateCertAction}))
}

func TestVersionCheckedButNotInitialized(t *testing.T) {
	// Setup fake client
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()
	director := actor.ClusterDirector{}

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")
	cluster.SetTrue(api.CrdbVersionChecked)

	actions := director.GetActionsToExecute(cluster)
	require.True(t, cmp.Equal(actions, []api.ActionType{api.GenerateCertAction, api.DeployAction, api.InitializeAction, api.ClusterRestartAction}))
}

func TestVersionCheckedAndInitialized(t *testing.T) {
	// Setup fake client
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()
	director := actor.ClusterDirector{}

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")
	cluster.SetTrue(api.InitializedCondition)
	cluster.SetTrue(api.CrdbVersionChecked)

	actions := director.GetActionsToExecute(cluster)
	require.True(t, cmp.Equal(actions, []api.ActionType{api.DecommissionAction, api.PartialUpdateAction, api.ResizePVCAction, api.DeployAction, api.ClusterRestartAction}))
}
