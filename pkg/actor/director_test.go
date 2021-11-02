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
	"fmt"
	"testing"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
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
	clientset := fake.NewSimpleClientset()
	director := actor.NewDirector(scheme, client, nil, clientset)

	return cluster, director
}

func TestDecommissionFeatureGate(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	cluster.SetTrue(api.CrdbInitializedCondition)

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true"))
	actors := director.GetActorsToExecute(cluster)
	require.True(t, containsAction(actors, api.DecommissionAction))

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=false"))
	actors = director.GetActorsToExecute(cluster)
	require.False(t, containsAction(actors, api.DecommissionAction))
}

func TestVersionValidatorFeatureGate(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	cluster.SetTrue(api.CrdbInitializedCondition)

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("CrdbVersionValidator=true"))
	actors := director.GetActorsToExecute(cluster)
	require.True(t, containsAction(actors, api.VersionCheckerAction))

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("CrdbVersionValidator=false"))
	actors = director.GetActorsToExecute(cluster)
	require.False(t, containsAction(actors, api.VersionCheckerAction))
}

func TestResizePVCFeatureGate(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	cluster.SetTrue(api.CrdbInitializedCondition)

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("ResizePVC=true"))
	actors := director.GetActorsToExecute(cluster)
	require.True(t, containsAction(actors, api.ResizePVCAction))

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("ResizePVC=false"))
	actors = director.GetActorsToExecute(cluster)
	require.False(t, containsAction(actors, api.ResizePVCAction))
}

func TestClusterRestartFeatureGate(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)

	cluster.SetTrue(api.CrdbInitializedCondition)
	cluster.SetTrue(api.CrdbVersionChecked)

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("ClusterRestart=true"))
	actors := director.GetActorsToExecute(cluster)
	require.True(t, containsAction(actors, api.ClusterRestartAction))

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("ClusterRestart=false"))
	actors = director.GetActorsToExecute(cluster)
	require.False(t, containsAction(actors, api.ClusterRestartAction))
}

func actorTypes(actors []actor.Actor) []api.ActionType {
	types := make([]api.ActionType, 0, len(actors))
	for _, a := range actors {
		types = append(types, a.GetActionType())
	}
	return types
}

func TestAllConditionCombinations(t *testing.T) {
	cluster, director := createTestDirectorAndCluster(t)
	err := utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")
	require.NoError(t, err)

	tests := []struct {
		trueConditions []api.ClusterConditionType
		expectedActors []api.ActionType
	}{
		{
			trueConditions: []api.ClusterConditionType{},
			expectedActors: []api.ActionType{api.VersionCheckerAction, api.RequestCertAction},
		},
		{
			trueConditions: []api.ClusterConditionType{api.CrdbInitializedCondition},
			expectedActors: []api.ActionType{api.DecommissionAction, api.VersionCheckerAction, api.RequestCertAction, api.ResizePVCAction},
		},
		{
			trueConditions: []api.ClusterConditionType{api.CertificateGenerated},
			expectedActors: []api.ActionType{api.VersionCheckerAction},
		},
		{
			trueConditions: []api.ClusterConditionType{api.CrdbVersionChecked},
			expectedActors: []api.ActionType{api.RequestCertAction, api.DeployAction, api.InitializeAction, api.ClusterRestartAction},
		},
		{
			trueConditions: []api.ClusterConditionType{api.CrdbInitializedCondition, api.CertificateGenerated},
			expectedActors: []api.ActionType{api.DecommissionAction, api.VersionCheckerAction, api.ResizePVCAction},
		},
		{
			trueConditions: []api.ClusterConditionType{api.CrdbInitializedCondition, api.CrdbVersionChecked},
			expectedActors: []api.ActionType{api.DecommissionAction, api.RequestCertAction, api.PartitionedUpdateAction, api.ResizePVCAction, api.DeployAction, api.ClusterRestartAction},
		},
		{
			trueConditions: []api.ClusterConditionType{api.CertificateGenerated, api.CrdbVersionChecked},
			expectedActors: []api.ActionType{api.DeployAction, api.InitializeAction, api.ClusterRestartAction},
		},
		{
			trueConditions: []api.ClusterConditionType{api.CrdbInitializedCondition, api.CertificateGenerated, api.CrdbVersionChecked},
			expectedActors: []api.ActionType{api.DecommissionAction, api.PartitionedUpdateAction, api.ResizePVCAction, api.DeployAction, api.ClusterRestartAction},
		},
	}

	for _, test := range tests {
		cluster.SetFalse(api.CrdbInitializedCondition)
		cluster.SetFalse(api.CertificateGenerated)
		cluster.SetFalse(api.CrdbVersionChecked)
		for _, c := range test.trueConditions {
			cluster.SetTrue(c)
		}

		actors := director.GetActorsToExecute(cluster)
		expActors := append([]api.ActionType{api.SetupRBACAction}, test.expectedActors...)
		require.Equal(t, expActors, actorTypes(actors), fmt.Sprintf("true conditions: %v", test.trueConditions))
	}
}
