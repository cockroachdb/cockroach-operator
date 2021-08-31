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
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/stretchr/testify/require"
	"testing"
)

func containsAction(actors []actor.Actor, action api.ActionType) bool {
	for _, a := range actors {
		if a.GetActionType() == action {
			return true
		}
	}
	return false
}

func createTestDirectorAndCluster(t *testing.T) (*resource.Cluster, actor.Director) {
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()

	scheme := testutil.InitScheme(t)
	client := testutil.NewFakeClient(scheme)
	director := actor.NewDirector(scheme, client, nil)

	return cluster, director
}

func TestDecommissionFeatureGate(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	cluster.SetTrue(api.InitializedCondition)

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true")
	actors := director.GetActorsToExecute(cluster)
	require.True(t, containsAction(actors, api.DecommissionAction))

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=false")
	actors = director.GetActorsToExecute(cluster)
	require.False(t, containsAction(actors, api.DecommissionAction))
}

func TestVersionValidatorFeatureGate(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	cluster.SetTrue(api.InitializedCondition)

	utilfeature.DefaultMutableFeatureGate.Set("CrdbVersionValidator=true")
	actors := director.GetActorsToExecute(cluster)
	require.True(t, containsAction(actors, api.VersionCheckerAction))

	utilfeature.DefaultMutableFeatureGate.Set("CrdbVersionValidator=false")
	actors = director.GetActorsToExecute(cluster)
	require.False(t, containsAction(actors, api.VersionCheckerAction))
}

func TestResizePVCFeatureGate(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	cluster.SetTrue(api.InitializedCondition)

	utilfeature.DefaultMutableFeatureGate.Set("ResizePVC=true")
	actors := director.GetActorsToExecute(cluster)
	require.True(t, containsAction(actors, api.ResizePVCAction))

	utilfeature.DefaultMutableFeatureGate.Set("ResizePVC=false")
	actors = director.GetActorsToExecute(cluster)
	require.False(t, containsAction(actors, api.ResizePVCAction))
}

func TestClusterRestartFeatureGate(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	cluster.SetTrue(api.InitializedCondition)
	cluster.SetTrue(api.CrdbVersionChecked)

	utilfeature.DefaultMutableFeatureGate.Set("ClusterRestart=true")
	actors := director.GetActorsToExecute(cluster)
	require.True(t, containsAction(actors, api.ClusterRestartAction))

	utilfeature.DefaultMutableFeatureGate.Set("ClusterRestart=false")
	actors = director.GetActorsToExecute(cluster)
	require.False(t, containsAction(actors, api.ClusterRestartAction))
}

func actorsHaveTypes(actors []actor.Actor, actionTypes []api.ActionType) bool {
	if len(actors) != len(actionTypes) {
		return false
	}
	for i, a := range actors {
		if a.GetActionType() != actionTypes[i] {
			return false
		}
	}
	return true
}

func TestTotallyUninitialized(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")

	actors := director.GetActorsToExecute(cluster)
	require.True(t, actorsHaveTypes(actors, []api.ActionType{api.VersionCheckerAction, api.RequestCertAction}))
}

func TestVersionCheckedButNotInitialized(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")
	cluster.SetTrue(api.CrdbVersionChecked)

	actors := director.GetActorsToExecute(cluster)
	require.True(t, actorsHaveTypes(actors, []api.ActionType{api.RequestCertAction, api.DeployAction, api.InitializeAction, api.ClusterRestartAction}))
}

func TestInitializedButNotVersionChecked(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")
	cluster.SetTrue(api.InitializedCondition)

	actors := director.GetActorsToExecute(cluster)
	require.True(t, actorsHaveTypes(actors, []api.ActionType{api.DecommissionAction, api.VersionCheckerAction, api.ResizePVCAction}))
}

func TestVersionCheckedAndInitialized(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")
	cluster.SetTrue(api.InitializedCondition)
	cluster.SetTrue(api.CrdbVersionChecked)

	actors := director.GetActorsToExecute(cluster)
	require.True(t, actorsHaveTypes(actors, []api.ActionType{api.DecommissionAction, api.PartialUpdateAction, api.ResizePVCAction, api.DeployAction, api.ClusterRestartAction}))
}
