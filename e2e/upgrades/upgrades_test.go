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

package upgrades

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	testenv "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TODO parallel seems to be buggy.  Not certain why, but we need to figure out if running with the operator
// deployed in the cluster helps
// We may have a threadsafe problem where one test starts messing with another test
var parallel = *flag.Bool("parallel", false, "run tests in parallel")

// TODO should we make this an atomic that is created by evn pkg?
var env *testenv.ActiveEnv

// TestMain wraps the unit tests. Set TEST_DO_NOT_USE_KIND evnvironment variable to any value
// if you do not want this test to start a k8s cluster using kind.
func TestMain(m *testing.M) {
	e := testenv.CreateActiveEnvForTest([]string{"..", ".."})
	env = e.Start()
	code := testenv.RunCode(m, e)
	os.Exit(code)
}

// TestUpgradesMinorVersion tests a minor version bump
func TestUpgradesMinorVersion(t *testing.T) {

	// We are testing a Minor Version Upgrade with
	// partition update
	// Going from v20.2.8 to v20.2.9

	if parallel {
		t.Parallel()
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.8").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			Name: "upgrades the cluster to the next patch version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = "cockroachdb/cockroach:v20.2.9"
				require.NoError(t, sb.Update(current))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDbContainersToUseImage(t, sb, current)
				t.Log("Done with upgrade")
			},
		},
	}

	steps.Run(t)
}

// TestUpgradesMajorVersion20to21 tests a major version upgrade
func TestUpgradesMajorVersion20to21(t *testing.T) {

	// We are doing a major version upgrade here
	// 20 to 21

	if parallel {
		t.Parallel()
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.9").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := testutil.Steps{
		{
			Name: "creates a 1-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			Name: "upgrades the cluster to the next minor version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = "cockroachdb/cockroach:v21.1.0"
				require.NoError(t, sb.Update(current))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDbContainersToUseImage(t, sb, current)
				t.Log("Done with major upgrade")
			},
		},
	}

	steps.Run(t)
}

// TestUpgradesMajorVersion20_1To20_2 is another major version upgrade
func TestUpgradesMajorVersion20_1To20_2(t *testing.T) {

	if parallel {
		t.Parallel()
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage("cockroachdb/cockroach:v20.1.16").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			Name: "upgrades the cluster to the next major version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = "cockroachdb/cockroach:v20.2.10"
				require.NoError(t, sb.Update(current))
				// we wait 10 min because we will be waiting 3 min for each pod because
				// v20.1.16 does not have curl installed
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDbContainersToUseImage(t, sb, current)
				t.Log("Done with major upgrade")
			},
		},
		{
			Name: "test downgrade set",
			Test: func(t *testing.T) {
				testutil.RequireDownGradeOptionSet(t, sb, builder, "20.1")
				t.Log("downgrade option set")
			},
		},
	}

	steps.Run(t)
}
