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
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func createTestDirectorAndStableCluster(t *testing.T) (*resource.Cluster, actor.Director) {
	var numNodes int32 = 4
	version := "fake.version"
	storage := "1Gi"

	// Initialize a mock cluster with various attributes
	clusterAnnotations := map[string]string{
		resource.CrdbVersionAnnotation: version,
	}
	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore(storage, "standard" /* default storage class in KIND */).
		WithNodeCount(numNodes).
		WithClusterAnnotations(clusterAnnotations).
		Cluster()
	// A stable cluster has a checked version and is initialized
	cluster.SetTrue(api.CrdbVersionChecked)
	cluster.SetTrue(api.CrdbInitializedCondition)

	// Mock node for our mock cluster
	node := &v1.Node{}
	objs := []runtime.Object{
		node,
	}

	// Mock components of our mock cluster
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

	quantity, _ := apiresource.ParseQuantity(storage)
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb",
			Namespace: "default",
			Annotations: map[string]string{
				resource.CrdbVersionAnnotation: version,
			},
		},
		Status: appsv1.StatefulSetStatus{
			Replicas:        numNodes,
			CurrentReplicas: numNodes,
		},
		Spec: appsv1.StatefulSetSpec{
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: quantity,
							},
						},
					},
				},
			},
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

	// Build up all the required pieces of our cluster components
	l := labels.Common(cluster.Unwrap())
	kd, _ := kube.NewKubernetesDistribution().Get(context.Background(), fake.NewSimpleClientset(objs...), zapr.NewLogger(zaptest.NewLogger(t)))
	kd = "kubernetes-operator-" + kd
	scheme := testutil.InitScheme(t)
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

	// Construct mock cluster access
	client := testutil.NewFakeClient(scheme, objs...)
	clientset := fake.NewSimpleClientset(objs...)
	config := &rest.Config{}
	director := actor.NewDirector(scheme, client, config, clientset)

	return cluster, director
}

func TestNoActionRequired(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)

	actor, err := director.GetActorToExecute(context.Background(), cluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, nil, actor)
}

func TestNeedsRestart(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)
	updated := cluster.Unwrap()

	// Trigger restart by adding restart annotation
	updated.Annotations = make(map[string]string)
	updated.Annotations[resource.CrdbRestartTypeAnnotation] = "Rolling"

	newCluster := resource.NewCluster(updated)
	actor, err := director.GetActorToExecute(context.Background(), &newCluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, api.ClusterRestartAction, actor.GetActionType())
}

func TestNeedsDecommission(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)
	updated := cluster.Unwrap()

	// Trigger decommission by decreasing nodes
	updated.Spec.Nodes = 3

	newCluster := resource.NewCluster(updated)
	actor, err := director.GetActorToExecute(context.Background(), &newCluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, api.DecommissionAction, actor.GetActionType())
}

func TestNeedsVersionCheck(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)

	// Trigger version check by setting condition to false
	cluster.SetFalse(api.CrdbVersionChecked)

	actor, err := director.GetActorToExecute(context.Background(), cluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, api.VersionCheckerAction, actor.GetActionType())
}

func TestNeedsCertificate(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)
	updated := cluster.Unwrap()

	// Trigger certificate generation by enabling TLS
	updated.Spec.TLSEnabled = true

	newCluster := resource.NewCluster(updated)
	actor, err := director.GetActorToExecute(context.Background(), &newCluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, api.GenerateCertAction, actor.GetActionType())
}

func TestNeedsUpdate(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)
	updated := cluster.Unwrap()

	// Trigger update by changing requested version
	updated.Annotations = make(map[string]string)
	updated.Annotations[resource.CrdbVersionAnnotation] = "fake.version.2"

	newCluster := resource.NewCluster(updated)
	actor, err := director.GetActorToExecute(context.Background(), &newCluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, api.PartitionedUpdateAction, actor.GetActionType())
}

func TestNeedsPVCResize(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)
	updated := cluster.Unwrap()

	// Trigger PVC resize by increasing requested amount
	quantity, _ := apiresource.ParseQuantity("2Gi")
	updated.Spec.DataStore.VolumeClaim.PersistentVolumeClaimSpec.Resources.Requests[v1.ResourceStorage] = quantity

	newCluster := resource.NewCluster(updated)
	actor, err := director.GetActorToExecute(context.Background(), &newCluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, api.ResizePVCAction, actor.GetActionType())
}

func TestNeedsDeploy(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)
	updated := cluster.Unwrap()

	// Trigger decommission by increasing nodes
	updated.Spec.Nodes = 5

	newCluster := resource.NewCluster(updated)
	actor, err := director.GetActorToExecute(context.Background(), &newCluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, api.DeployAction, actor.GetActionType())
}

func TestNeedsInitialization(t *testing.T) {
	cluster, director := createTestDirectorAndStableCluster(t)

	// Trigger version check by setting condition to false
	cluster.SetFalse(api.CrdbInitializedCondition)

	actor, err := director.GetActorToExecute(context.Background(), cluster, zapr.NewLogger(zaptest.NewLogger(t)))
	require.Nil(t, err)
	require.Equal(t, api.InitializeAction, actor.GetActionType())
}

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
