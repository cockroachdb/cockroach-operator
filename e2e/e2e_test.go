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

package e2e

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	testenv "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil/paths"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

// TODO parallel seems to be buggy.  Not certain why, but we need to figure out if running with the operator
// deployed in the cluster helps
// We may have a threadsafe problem where one test starts messing with another test
var parallel = *flag.Bool("parallel", false, "run tests in parallel")

// run the pvc test
var pvc = flag.Bool("pvc", false, "run pvc test")

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

// TestMain wraps the unit tests. Set TEST_DO_NOT_USE_KIND evnvironment variable to any value
// if you do not want this test to start a k8s cluster using kind.
func TestMain(m *testing.M) {
	flag.Parse()

	os.Setenv("USE_EXISTING_CLUSTER", "true")
	paths.MaybeSetEnv("PATH", "kubetest2-kind", "hack", "bin", "kubetest2-kind")

	e := testenv.NewEnv(runtime.NewSchemeBuilder(api.AddToScheme),
		filepath.Join("..", "config", "crd", "bases"),
		filepath.Join("..", "config", "rbac", "bases"))

	env = e.Start()
	code := m.Run()
	e.Stop()
	os.Exit(code)
}

func TestCreatesSecureCluster(t *testing.T) {

	// Test Creating a secure cluster
	// No actions on the cluster just create it and
	// tear it down.

	if parallel {
		t.Parallel()
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	paths.MaybeSetEnv("PATH", "kubetest2-kind", "hack", "bin", "kubetest2-kind")

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	create := Step{
		name: "creates 3-node secure cluster",
		test: func(t *testing.T) {
			require.NoError(t, sb.Create(builder.Cr()))

			RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			requireDatabaseToFunction(t, sb, builder)
			t.Log("Done with basic cluster")
		},
	}

	steps := Steps{create}

	steps.Run(t)
}

func TestUpgradesMinorVersion(t *testing.T) {

	// We are testing a Minor Version Upgrade with
	// partition update
	// Going from 19.2.5 to 19.2.6

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
		WithImage("cockroachdb/cockroach:v19.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := Steps{
		{
			name: "creates a 1-node secure cluster",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			name: "upgrades the cluster to the next patch version",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = "cockroachdb/cockroach:v19.2.6"
				require.NoError(t, sb.Update(current))

				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				requireDbContainersToUseImage(t, sb, current)
				t.Log("Done with upgrade")
			},
		},
	}

	steps.Run(t)
}

func TestUpgradesMajorVersion19to20(t *testing.T) {

	// We are doing a major version upgrade here
	// 19 to 20

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
		WithImage("cockroachdb/cockroach:v19.2.6").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := Steps{
		{
			name: "creates a 1-node secure cluster",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			name: "upgrades the cluster to the next minor version",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = "cockroachdb/cockroach:v20.1.1"
				require.NoError(t, sb.Update(current))

				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				requireDbContainersToUseImage(t, sb, current)
				t.Log("Done with major upgrade")
			},
		},
	}

	steps.Run(t)
}

func TestUpgradesMajorVersion20_1To20_2(t *testing.T) {

	// Major version upgrade 20_1 to 20_2

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
		WithImage("cockroachdb/cockroach:v20.1.12").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := Steps{
		{
			name: "creates a 3-node secure cluster",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			name: "upgrades the cluster to the next minor version",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = "cockroachdb/cockroach:v20.2.5"
				require.NoError(t, sb.Update(current))

				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				requireDbContainersToUseImage(t, sb, current)
				t.Log("Done with major upgrade")
			},
		},
	}

	steps.Run(t)
}

func TestDecommissionFunctionality(t *testing.T) {

	// Testing removing and decommisioning a node.  We start at 4 node and then
	// remove the 4th node

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Does not seem to like running in parallel
	if parallel {
		t.Parallel()
	}
	testLog := zapr.NewLogger(zaptest.NewLogger(t))
	actor.Log = testLog
	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))
	builder := testutil.NewBuilder("crdb").WithNodeCount(4).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)
	steps := Steps{
		{
			name: "creates a 4-node secure cluster and tests db",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			name: "decommission a node",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Nodes = 3
				require.NoError(t, sb.Update(current))
				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				requireDecommissionNode(t, sb, builder, 3)
				requireDatabaseToFunction(t, sb, builder)
				t.Log("Done with decommision")
			},
		},
	}
	steps.Run(t)
}

func TestPVCResize(t *testing.T) {

	// Testing PVCResize
	if !*pvc {
		t.Skip("platform does not support pvc resize")
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	if parallel {
		t.Parallel()
	}
	testLog := zapr.NewLogger(zaptest.NewLogger(t))
	actor.Log = testLog
	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))
	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)
	steps := Steps{
		{
			name: "creates a 3-node secure cluster db",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			name: "resize PVC",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))
				quantity := apiresource.MustParse("2Gi")
				current.Spec.DataStore.VolumeClaim.PersistentVolumeClaimSpec.Resources.Requests[corev1.ResourceStorage] = quantity
				require.NoError(t, sb.Update(current))
				t.Log("updated CR")

				requirePVCToResize(t, sb, builder, quantity)
				t.Log("here resized")
			},
		},
	}
	steps.Run(t)
}
func TestRollingRestart(t *testing.T) {

	// Restart cluster nodes in a rolling cadence, one after the other

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
		WithImage("cockroachdb/cockroach:v20.2.5").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := Steps{
		{
			name: "creates a 3-node secure cluster",
			test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
			},
		},
		{
			name: "restarts the cluster using the rolling functionality",
			test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				pods, err := fetchPodsInStatefulSet(sb, labels.Common(current).Selector())
				require.NoError(t, err)
				require.NotNil(t, pods)

				podCreationTimes := make(map[string]metav1.Time)
				for _, pod := range pods {
					if strings.Contains(pod.Name, "-vcheck-") {
						continue // Skip the validate version pod
					}
					podCreationTimes[pod.Name] = pod.CreationTimestamp
				}

				current.Annotations["crdb.io/restarttype"] = "rolling"
				require.NoError(t, sb.Update(current))

				// Need to poll for getting rid of the restart annotation
				err = wait.Poll(10*time.Second, 500*time.Second, func() (done bool, err error) {
					current = builder.Cr()
					if err = sb.Get(current); err != nil {
						return false, err
					}

					if rt, ok := current.Annotations["crdb.io/restarttype"]; !ok {
						return true, nil
					} else if rt == "" { // The annotation was deleted
						return true, nil
					}

					return false, nil
				})
				require.NoError(t, err)

				// Make sure the cluster is in a good state after the restart is finished
				RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)

				pods, err = fetchPodsInStatefulSet(sb, labels.Common(current).Selector())
				require.NoError(t, err)
				require.NotNil(t, pods)

				podRestartTimes := make(map[string]metav1.Time)
				for _, pod := range pods {
					if strings.Contains(pod.Name, "-vcheck-") {
						continue // Skip the validate version pod
					}
					podRestartTimes[pod.Name] = pod.CreationTimestamp
				}

				for name, ts := range podCreationTimes {
					restartTS, ok := podRestartTimes[name]
					require.True(t, ok, "Unable to find pod %s", name)
					require.True(t, ts.Before(&restartTS), "Pod %s did not restart %v not before %v", name, ts, restartTS)
				}

				t.Log("Done with rolling restart")
			},
		},
	}

	steps.Run(t)
}
