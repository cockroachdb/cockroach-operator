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

//func TestActorOrder(t *testing.T) {
//
//	// So the actor order is so important that we need to ensure that the slice
//	// does not get touched.
//	// This test ensures that the name of the actor is in the correct order.
//
//	// Setup fake client
//	builder := fake.NewClientBuilder()
//
//	client := builder.Build()
//	actors := NewOperatorActions(nil, client, nil)
//
//	actorNamesInOrder := []string{
//		"*actor.decommission",
//		"*actor.versionChecker",
//		"*actor.generateCert",
//		"*actor.partitionedUpdate",
//		"*actor.resizePVC",
//		"*actor.deploy",
//		"*actor.initialize",
//		"*actor.clusterRestart",
//	}
//
//	// TODO this seems kinda hacky, but I could not get a.(type) working
//	// so I am just testing the name of the string. Maybe I could test the
//	// struct type, but hey this works for now.
//	for i, a := range actors {
//		xType := reflect.TypeOf(a)
//		expected := actorNamesInOrder[i]
//		require.True(t, xType.String() == expected, "expected:", expected, "got", xType.String())
//	}
//
//}

//func TestFakeTest(t *testing.T) {
//	// Setup fake client
//	cluster := testutil.NewBuilder("cockroachdb").
//		Namespaced("default").
//		WithUID("cockroachdb-uid").
//		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
//		WithNodeCount(1).Cluster()
//	_ = cluster
//
//	require.True(t, false)
//}
//
//func TestDecommissionFeatureGate(t *testing.T) {
//	// Setup fake client
//	director := actor.ClusterDirector{}
//
//	utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
//}
