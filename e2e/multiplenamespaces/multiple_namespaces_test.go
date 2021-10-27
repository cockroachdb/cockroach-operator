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

package multiplenamespaces

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


// TestCreateCrdbClusterInMultipleNamespaces creates two clusters of 4 nodes in different namespace
// and then decommissions one of the CRDB nodes in the first CRDB cluster.
// It then checks that the cluster is stable and that decommissioning is successful.
func TestCreateCrdbClusterInMultipleNamespaces(t *testing.T) {

	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("AutoPrunePVC=true"))
	require.NoError(t, utilfeature.DefaultMutableFeatureGate.Set("MultipleNamespaces=true"))

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

	sb := testenv.NewDiffingSandboxWithoutNamespace(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := createBuilder(t, env)
	builder2 := createBuilder(t, env)
	steps := testutil.Steps{
		createBuildCrdbStep(sb, builder),
		createBuildCrdbStep(sb, builder2),
		{
			Name: "decommission a node with pvc pruner in namespace: " + builder.Namespace(),
			Test: func(t *testing.T) {
				ns := builder.Namespace()
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Nodes = 3
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeoutWithNamespace(t, sb, builder, ns, 500*time.Second)
				testutil.RequireDecommissionNodeWithNamespace(t, sb, builder, ns, 3)
				testutil.RequireDatabaseToFunctionWithNamespace(t, sb, builder, ns)
				t.Log("Done with decommission")
				testutil.RequireNumberOfPVCsWithNamespace(t, context.TODO(), sb, builder, ns, 3)
			},
		},
	}
	steps.Run(t)
}

func createBuildCrdbStep(sb testenv.DiffingSandbox, builder testutil.ClusterBuilder) testutil.Step {
	return testutil.Step {
		Name: "creates a 4-node secure cluster and tests db in namespace: " + builder.Namespace(),
		Test: func(t *testing.T) {
			ns := builder.Namespace()
			require.NoError(t, sb.Create(builder.Cr()))
			testutil.RequireClusterToBeReadyEventuallyTimeoutWithNamespace(t, sb, builder, ns, 500*time.Second)
			testutil.RequireNumberOfPVCsWithNamespace(t, context.TODO(), sb, builder, ns, 4)
		},
	}
}

func createBuilder(t *testing.T, env *testenv.ActiveEnv) testutil.ClusterBuilder {
	ns := testenv.CreateNamespace(t, env, "crdb-")
	if err := testenv.CreateAndBindServiceAccountNamespaced(env, ns); err != nil {
		t.Fatal(err)
	}
	builder := testutil.NewBuilder("crdb").Namespaced(ns).WithNodeCount(4).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)
	return builder
}
