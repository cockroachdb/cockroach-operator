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
	"context"
	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func createTestDirectorAndStableCluster(t *testing.T) (*resource.Cluster, actor.Director) {
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(4).Cluster()
	scheme := testutil.InitScheme(t)

	// Mock node for our mock cluster
	node := &v1.Node{}
	objs := []runtime.Object{
		node,
	}

	//
	discoveryService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb",
			Namespace: "default",
		},
	}
	publicService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb-public",
			Namespace: "default",
		},
	}
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb",
			Namespace: "default",
		},
	}
	podDisruptionBudget := &v1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb",
			Namespace: "default",
		},
	}
	components := []client.Object{
		discoveryService,
		publicService,
		statefulSet,
		podDisruptionBudget,
	}

	l := labels.Common(cluster.Unwrap())
	kd, _ := kube.NewKubernetesDistribution().Get(context.Background(), fake.NewSimpleClientset(objs...), zapr.NewLogger(zaptest.NewLogger(t)))
	kd = "kubernetes-operator-" + kd
	builders := []resource.Builder{
		resource.DiscoveryServiceBuilder{Cluster: cluster, Selector: l.Selector(nil)},
		resource.PublicServiceBuilder{Cluster: cluster, Selector: l.Selector(nil)},
		resource.StatefulSetBuilder{Cluster: cluster, Selector: l.Selector(nil), Telemetry: kd},
		resource.PdbBuilder{Cluster: cluster, Selector: l.Selector(nil)},
	}
	for i := range builders {
		resource.Reconciler{
			ManagedResource: resource.ManagedResource{
				Labels: l,
			},
			Builder: builders[i],

			Owner:  cluster.Unwrap(),
			Scheme: scheme,
		}.CompleteBuild(components[i].DeepCopyObject(), components[i])
		objs = append(objs, components[i])
	}

	// TODO: need to also construct other three things: public service, stateful set, pod disruption budget

	client := testutil.NewFakeClient(scheme, objs...)
	clientset := fake.NewSimpleClientset(objs...)
	config := &rest.Config{}
	director := actor.NewDirector(scheme, client, config, clientset)

	return cluster, director
}

func TestNoActionRequired(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)
	cluster.SetTrue(api.CrdbVersionChecked)
	cluster.SetTrue(api.CrdbInitializedCondition)

	actor, err := director.GetActorToExecute(context.Background(), cluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, nil, actor)
}

//func TestNeedsRestart(t *testing.T) {
//	cluster, director := createTestDirectorAndCluster(t)
//
//	actor, err := director.GetActorToExecute(context.Background(), cluster, nil)
//	require.Nil(t, err)
//	require.Equal(t, api.ClusterRestartAction, actor.GetActionType())
//}

//func TestDecommissionFeatureGate(t *testing.T) {
//	cluster, director := createTestDirectorAndCluster(t)
//
//	cluster.SetTrue(api.CrdbInitializedCondition)
//
//	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true")
//	actors := director.GetActorsToExecute(cluster)
//	require.True(t, containsAction(actors, api.DecommissionAction))
//
//	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=false")
//	actors = director.GetActorsToExecute(cluster)
//	require.False(t, containsAction(actors, api.DecommissionAction))
//}
//
//func TestVersionValidatorFeatureGate(t *testing.T) {
//	cluster, director := createTestDirectorAndCluster(t)
//
//	cluster.SetTrue(api.CrdbInitializedCondition)
//
//	utilfeature.DefaultMutableFeatureGate.Set("CrdbVersionValidator=true")
//	actors := director.GetActorsToExecute(cluster)
//	require.True(t, containsAction(actors, api.VersionCheckerAction))
//
//	utilfeature.DefaultMutableFeatureGate.Set("CrdbVersionValidator=false")
//	actors = director.GetActorsToExecute(cluster)
//	require.False(t, containsAction(actors, api.VersionCheckerAction))
//}
//
//func TestResizePVCFeatureGate(t *testing.T) {
//	cluster, director := createTestDirectorAndCluster(t)
//
//	cluster.SetTrue(api.CrdbInitializedCondition)
//
//	utilfeature.DefaultMutableFeatureGate.Set("ResizePVC=true")
//	actors := director.GetActorsToExecute(cluster)
//	require.True(t, containsAction(actors, api.ResizePVCAction))
//
//	utilfeature.DefaultMutableFeatureGate.Set("ResizePVC=false")
//	actors = director.GetActorsToExecute(cluster)
//	require.False(t, containsAction(actors, api.ResizePVCAction))
//}
//
//func TestClusterRestartFeatureGate(t *testing.T) {
//	cluster, director := createTestDirectorAndCluster(t)
//
//	cluster.SetTrue(api.CrdbInitializedCondition)
//	cluster.SetTrue(api.CrdbVersionChecked)
//
//	utilfeature.DefaultMutableFeatureGate.Set("ClusterRestart=true")
//	actors := director.GetActorsToExecute(cluster)
//	require.True(t, containsAction(actors, api.ClusterRestartAction))
//
//	utilfeature.DefaultMutableFeatureGate.Set("ClusterRestart=false")
//	actors = director.GetActorsToExecute(cluster)
//	require.False(t, containsAction(actors, api.ClusterRestartAction))
//}
//
//func actorTypes(actors []actor.Actor) []api.ActionType {
//	types := make([]api.ActionType, 0, len(actors))
//	for _, a := range actors {
//		types = append(types, a.GetActionType())
//	}
//	return types
//}
//
//func TestAllConditionCombinations(t *testing.T) {
//	cluster, director := createTestDirectorAndCluster(t)
//	utilfeature.DefaultMutableFeatureGate.Set("UseDecommission=true,CrdbVersionValidator=true,ResizePVC=true,ClusterRestart=true")
//
//	tests := []struct {
//		trueConditions []api.ClusterConditionType
//		expectedActors []api.ActionType
//	}{
//		{
//			trueConditions: []api.ClusterConditionType{},
//			expectedActors: []api.ActionType{api.VersionCheckerAction, api.RequestCertAction},
//		},
//		{
//			trueConditions: []api.ClusterConditionType{api.CrdbInitializedCondition},
//			expectedActors: []api.ActionType{api.DecommissionAction, api.VersionCheckerAction, api.RequestCertAction, api.ResizePVCAction},
//		},
//		{
//			trueConditions: []api.ClusterConditionType{api.CertificateGenerated},
//			expectedActors: []api.ActionType{api.VersionCheckerAction},
//		},
//		{
//			trueConditions: []api.ClusterConditionType{api.CrdbVersionChecked},
//			expectedActors: []api.ActionType{api.RequestCertAction, api.DeployAction, api.InitializeAction, api.ClusterRestartAction},
//		},
//		{
//			trueConditions: []api.ClusterConditionType{api.CrdbInitializedCondition, api.CertificateGenerated},
//			expectedActors: []api.ActionType{api.DecommissionAction, api.VersionCheckerAction, api.ResizePVCAction},
//		},
//		{
//			trueConditions: []api.ClusterConditionType{api.CrdbInitializedCondition, api.CrdbVersionChecked},
//			expectedActors: []api.ActionType{api.DecommissionAction, api.RequestCertAction, api.PartitionedUpdateAction, api.ResizePVCAction, api.DeployAction, api.ClusterRestartAction},
//		},
//		{
//			trueConditions: []api.ClusterConditionType{api.CertificateGenerated, api.CrdbVersionChecked},
//			expectedActors: []api.ActionType{api.DeployAction, api.InitializeAction, api.ClusterRestartAction},
//		},
//		{
//			trueConditions: []api.ClusterConditionType{api.CrdbInitializedCondition, api.CertificateGenerated, api.CrdbVersionChecked},
//			expectedActors: []api.ActionType{api.DecommissionAction, api.PartitionedUpdateAction, api.ResizePVCAction, api.DeployAction, api.ClusterRestartAction},
//		},
//	}
//
//	for _, test := range tests {
//		cluster.SetFalse(api.CrdbInitializedCondition)
//		cluster.SetFalse(api.CertificateGenerated)
//		cluster.SetFalse(api.CrdbVersionChecked)
//		for _, c := range test.trueConditions {
//			cluster.SetTrue(c)
//		}
//
//		actors := director.GetActorsToExecute(cluster)
//		require.Equal(t, test.expectedActors, actorTypes(actors), fmt.Sprintf("true conditions: %v", test.trueConditions))
//	}
//}
