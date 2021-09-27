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
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// We cannot do this since we are creatin the RBAC components now
// for openshift.  We need to create those only once.
var parallel = *flag.Bool("parallel", false, "run tests in parallel")

// TODO move these into a common file
var MinorVersion1 string = "cockroachdb/cockroach:v20.2.8"
var MinorVersion2 string = "cockroachdb/cockroach:v20.2.9"
var MajorVersion string = "cockroachdb/cockroach:v21.1.0"

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

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage(MinorVersion1).
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

				updated := current.DeepCopy()
				updated.Spec.Image.Name = MinorVersion2
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
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

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage(MinorVersion2).
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

				updated := current.DeepCopy()
				updated.Spec.Image.Name = MajorVersion
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
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

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

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

				updated := current.DeepCopy()
				updated.Spec.Image.Name = "cockroachdb/cockroach:v20.2.10"
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))
				// we wait 10 min because we will be waiting 3 min for each pod because
				// v20.1.16 does not have curl installed
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
				t.Log("Done with major upgrade")
			},
		},
	}

	steps.Run(t)
}

// TestUpgradesMinorVersionThenRollback tests a minor version bump
// then rollsback that upgrade
func TestUpgradesMinorVersionThenRollback(t *testing.T) {

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

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage(MinorVersion1).
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

				updated := current.DeepCopy()
				updated.Spec.Image.Name = MinorVersion2
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
				t.Log("Done with upgrade")
			},
		},
		{
			Name: "downgrades the cluster to the old patch version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Image.Name = MinorVersion1
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
				t.Log("Done with downgrade")
			},
		},
	}

	steps.Run(t)
}
