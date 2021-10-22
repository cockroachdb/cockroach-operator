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
	"testing"
	"time"

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

var validImage = "cockroachdb/cockroach:v21.1.7"
var nonExistentImage = "cockroachdb/cockroach-non-existent:v21.1.999"
var invalidImage = "nginx:latest"
var crdbVersion = "v21.1.7"
var relatedImageEnvName = "RELATED_IMAGE_COCKROACH_v21_1_7"

// TestCreateInsecureCluster tests the creation of insecure cluster, and it should be successful.
func TestCreateInsecureCluster(t *testing.T) {
	// Test Creating an insecure cluster
	// No actions on the cluster just create it and
	// tear it down.

	if parallel {
		t.Parallel()
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).
		WithImage("cockroachdb/cockroach:v21.1.6").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := testutil.Steps{
		{
			Name: "creates 3-node insecure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDatabaseToFunctionInsecure(t, sb, builder)

				t.Log("Done with basic cluster")
			},
		},
	}
	steps.Run(t)
}

// TestCreatesSecureCluster tests the creation of secure cluster, and it should be successful.
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

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
		WithImage("cockroachdb/cockroach:v20.2.10").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

	steps := testutil.Steps{
		{
			Name: "creates 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, 500*time.Second)
				testutil.RequireDatabaseToFunction(t, sb, builder)
				t.Log("Done with basic cluster")
			},
		},
	}
	steps.Run(t)
}

// TestCreateSecureClusterWithInvalidVersion tests cluster creation with invalid version and the cluster should fail.
func TestCreateSecureClusterWithInvalidVersion(t *testing.T) {
	// Test create a cluster with invalid version
	// Check it went into ErrImagePull state and then marked CR into failed state
	// tear it down

	if parallel {
		t.Parallel()
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	testcases := []struct {
		imageVersion     string
		cockroachVersion string
	}{
		{
			nonExistentImage,
			"",
		},
		{
			validImage,
			crdbVersion,
		},
	}

	for _, testcase := range testcases {
		steps := testutil.Steps{
			{
				Name: "creates 3-node secure cluster with invalid image",
				Test: func(subT *testing.T) {
					e := testenv.CreateActiveEnvForTest()
					env := e.Start()

					sb := testenv.NewDiffingSandbox(subT, env)
					sb.StartManager(subT, controller.InitClusterReconcilerWithLogger(testLog))

					builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
						WithImage(testcase.imageVersion).
						WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)
					if testcase.cockroachVersion != "" {
						os.Setenv(relatedImageEnvName, nonExistentImage)
						builder = builder.WithCockroachDBVersion(testcase.cockroachVersion)
					}

					require.NoError(subT, sb.Create(builder.Cr()))
					testutil.RequireClusterInImagePullBackoff(subT, sb, builder)
					testutil.RequireClusterInFailedState(subT, sb, builder)
					subT.Log("Done with basic invalid cluster")
					require.NoError(subT, sb.Delete(builder.Cr()))
					e.Stop()
				},
			},
		}
		steps.Run(t)

	}
}

// TestCreateSecureClusterWithNonCRDBImage tests creating a cluster with non-valid image.
// Creation should fail and CR should be in failed state.
func TestCreateSecureClusterWithNonCRDBImage(t *testing.T) {
	// Test create a cluster with non valid CRDB image
	// Check it went into failed state in initialized state
	// tear it down

	if parallel {
		t.Parallel()
	}
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	testcases := []struct {
		imageVersion     string
		cockroachVersion string
	}{
		{
			invalidImage,
			"",
		},
		{
			validImage,
			crdbVersion,
		},
	}

	for _, testcase := range testcases {
		steps := testutil.Steps{
			{
				Name: "creates 3-node secure cluster with invalid image",
				Test: func(subT *testing.T) {
					e := testenv.CreateActiveEnvForTest()
					env := e.Start()

					sb := testenv.NewDiffingSandbox(subT, env)
					sb.StartManager(subT, controller.InitClusterReconcilerWithLogger(testLog))

					builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
						WithImage(testcase.imageVersion).
						WithPVDataStore("1Gi", "standard" /* default storage class in KIND */)

					if testcase.cockroachVersion != "" {
						os.Setenv(relatedImageEnvName, invalidImage)
						builder = builder.WithCockroachDBVersion(testcase.cockroachVersion)
					}

					require.NoError(subT, sb.Create(builder.Cr()))
					testutil.RequireClusterInFailedState(subT, sb, builder)
					subT.Log("Done with basic invalid cluster with image other than crdb")
					require.NoError(subT, sb.Delete(builder.Cr()))

					e.Stop()
				},
			},
		}
		steps.Run(t)
	}
}
