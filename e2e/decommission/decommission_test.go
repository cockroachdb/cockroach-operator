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

package decommission

import (
	"context"
	"flag"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	testenv "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TODO parallel seems to be buggy.  Not certain why, but we need to figure out if running with the operator
// deployed in the cluster helps
// We may have a threadsafe problem where one test starts messing with another test
var parallel = *flag.Bool("parallel", false, "run tests in parallel")

// TODO once prune pvc feature gate is set to "true" by default, we can
// remove this test.

// TestDecommissionFunctionalityWithPrune creates a cluster of 4 nodes and then decommissions on of the CRDB nodes.
// It then checks that the cluster is stable and that decommissioning is successful.
func TestDecommissionFunctionalityWithPrune(t *testing.T) {

	// Testing removing and decommissioning a node.  We start at 4 node and then
	// remove the 4th node

	// turn on featuregate since Decommission is disabled by default currently
	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("AutoPrunePVC=true"))

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Does not seem to like running in parallel
	if parallel {
		t.Parallel()
	}
	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))
	builder := testutil.NewBuilder("crdb").Namespaced(sb.Namespace).WithNodeCount(4).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)
	steps := testutil.Steps{
		{
			Name: "creates a 4-node secure cluster and tests db",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireNumberOfPVCs(t, context.TODO(), sb, builder, 4)
			},
		},
		{
			Name: "decommission a node with pvc pruner",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Nodes = 3
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDecommissionNode(t, sb, builder, 3)
				testutil.RequireDatabaseToFunction(t, sb, builder)
				t.Log("Done with decommission")
				testutil.RequireNumberOfPVCs(t, context.TODO(), sb, builder, 3)
			},
		},
	}
	steps.Run(t)
}

// TestDecommissionFunctionality creates a cluster of 4 nodes and then decommissions on of the CRDB nodes.
// It then checks that the cluster is stable and that decommissioning is successful.
func TestDecommissionFunctionality(t *testing.T) {

	// Testing removing and decommissioning a node.  We start at 4 node and then
	// remove the 4th node

	// making sure the feature gate is off for prunePVC
	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("AutoPrunePVC=false"))

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Does not seem to like running in parallel
	if parallel {
		t.Parallel()
	}
	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))
	builder := testutil.NewBuilder("crdb").Namespaced(sb.Namespace).WithNodeCount(4).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)
	steps := testutil.Steps{
		{
			Name: "creates a 4-node secure cluster and tests db",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireNumberOfPVCs(t, context.TODO(), sb, builder, 4)
			},
		},
		{
			Name: "decommission a node",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Nodes = 3
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDecommissionNode(t, sb, builder, 3)
				testutil.RequireDatabaseToFunction(t, sb, builder)
				t.Log("Done with decommission")

				testutil.RequireNumberOfPVCs(t, context.TODO(), sb, builder, 4)
			},
		},
	}
	steps.Run(t)
}
