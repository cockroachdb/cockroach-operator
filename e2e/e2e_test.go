/*
Copyright 2020 The Cockroach Authors

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

package e2e

import (
	"flag"
	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	testenv "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"k8s.io/apimachinery/pkg/runtime"
	"path/filepath"
	"testing"
)

var updateOpt = flag.Bool("update", false, "update the golden files of this test")

var env *testenv.ActiveEnv

type Step struct {
	name string
	test func(t *testing.T)
}

type Steps []Step

func (ss Steps) WithStep(s Step) Steps {
	return append(ss, s)
}

func (ss Steps) Run(t *testing.T) {
	for _, s := range ss {
		if !t.Run(s.name, s.test) {
			t.FailNow()
		}
	}
}

func TestMain(m *testing.M) {
	flag.Parse()

	e := testenv.NewEnv(runtime.NewSchemeBuilder(api.AddToScheme),
		filepath.Join("..", "config", "crd", "bases"))

	env = e.Start()

	e.StopAndExit(m.Run())
}

func TestCreatesInsecureCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	b := testutil.NewBuilder("crdb").WithNodeCount(1).WithEmptyDirDataStore()

	create := Step{
		name: "creates 1-node insecure cluster",
		test: func(t *testing.T) {
			require.NoError(t, sb.Create(b.Cr()))

			requireClusterToBeReadyEventually(t, sb, b)

			state, err := sb.Diff()
			require.NoError(t, err)

			expected := testutil.ReadOrUpdateGoldenFile(t, state, *updateOpt)

			testutil.AssertDiff(t, expected, state)
		},
	}

	steps := Steps{create}

	steps.Run(t)
}

func TestCreatesSecureClusterWithGeneratedCert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	b := testutil.NewBuilder("crdb").WithNodeCount(1).WithTLS().WithEmptyDirDataStore()

	create := Step{
		name: "creates 1-node secure cluster",
		test: func(t *testing.T) {
			require.NoError(t, sb.Create(b.Cr()))

			requireClusterToBeReadyEventually(t, sb, b)

			state, err := sb.Diff()
			require.NoError(t, err)

			expected := testutil.ReadOrUpdateGoldenFile(t, state, *updateOpt)

			testutil.AssertDiff(t, expected, state)
		},
	}

	steps := Steps{create}

	steps.Run(t)
}

func TestUpgradesMinorVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(1).WithTLS().
		WithImage("cockroachdb/cockroach:v19.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := Steps{
		{
			name: "creates a 1-node secure cluster",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				requireClusterToBeReadyEventually(t, sb, builder)
			},
		},
		{
			name: "upgrades the cluster to the next patch version",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image = "cockroachdb/cockroach:v19.2.6"
				require.NoError(t, sb.Update(current))

				requireClusterToBeReadyEventually(t, sb, builder)
				requireDbContainersToUseImage(t, sb, current)
			},
		},
	}

	steps.Run(t)
}

func TestUpgradesMajorVersion19to20(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(1).WithTLS().
		WithImage("cockroachdb/cockroach:v19.2.6").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := Steps{
		{
			name: "creates a 1-node secure cluster",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				requireClusterToBeReadyEventually(t, sb, builder)
			},
		},
		{
			name: "upgrades the cluster to the next minor version",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image = "cockroachdb/cockroach:v20.1.1"
				require.NoError(t, sb.Update(current))

				requireClusterToBeReadyEventually(t, sb, builder)
				requireDbContainersToUseImage(t, sb, current)
			},
		},
	}

	steps.Run(t)
}

func TestUpgradesMajorVersion19_1To19_2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(1).WithTLS().
		WithImage("cockroachdb/cockroach:v19.1.4").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := Steps{
		{
			name: "creates a 1-node secure cluster",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				requireClusterToBeReadyEventually(t, sb, builder)
			},
		},
		{
			name: "upgrades the cluster to the next minor version",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image = "cockroachdb/cockroach:v19.2.1"
				require.NoError(t, sb.Update(current))

				requireClusterToBeReadyEventually(t, sb, builder)
				requireDbContainersToUseImage(t, sb, current)
			},
		},
	}

	steps.Run(t)
}
