/*
Copyright 2026 The Cockroach Authors

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
	"encoding/json"
	"os"
	"testing"

	"github.com/cockroachdb/cockroach-operator/e2e"
	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	testenv "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var (
	crdbVersion         = "v21.2.7"
	relatedImageEnvName = "RELATED_IMAGE_COCKROACH_v21_2_7"
	validImage          = "cockroachdb/cockroach:v21.2.7"
)

// TestCreateInsecureCluster tests the creation of insecure cluster, and it should be successful.
func TestCreateInsecureCluster(t *testing.T) {
	// Test Creating an insecure cluster
	// No actions on the cluster just create it and
	// tear it down.

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	// Create cluster with different logging config than the default one.
	logJson := []byte(`{"sinks": {"file-groups": {"dev": {"channels": "DEV", "filter": "WARNING"}}}}`)
	logConfig := make(map[string]interface{})
	require.NoError(t, json.Unmarshal(logJson, &logConfig))
	testutil.RequireLoggingConfigMap(t, sb, "logging-configmap", string(logJson))

	builder := testutil.NewBuilder("crdb").WithNodeCount(3).
		WithImage(e2e.MajorVersion).WithClusterLogging("logging-configmap").
		WithPVDataStore("1Gi")

	steps := testutil.Steps{
		{
			Name: "creates 3-node insecure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
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
		WithImage(e2e.MajorVersion).
		WithPVDataStore("1Gi")

	steps := testutil.Steps{
		{
			Name: "creates 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
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

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	testcases := []struct {
		imageVersion     string
		cockroachVersion string
	}{
		{
			e2e.NonExistentVersion,
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
						WithPVDataStore("1Gi")
					if testcase.cockroachVersion != "" {
						os.Setenv(relatedImageEnvName, e2e.NonExistentVersion)
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

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	testcases := []struct {
		imageVersion     string
		cockroachVersion string
	}{
		{
			e2e.InvalidImage,
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
						WithPVDataStore("1Gi")

					if testcase.cockroachVersion != "" {
						os.Setenv(relatedImageEnvName, e2e.InvalidImage)
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

// TestCreateSecureClusterWithCRDBVersionSet tests creating a cluster with
// CRDBVersion set.
// Creation should succeed.
func TestCreateSecureClusterWithCRDBVersionSet(t *testing.T) {
	// Test create a cluster with non valid CRDB image
	// Check it went into failed state in initialized state
	// tear it down

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	steps := testutil.Steps{
		{
			Name: "creates 3-node secure cluster with valid version number",
			Test: func(subT *testing.T) {
				os.Setenv(relatedImageEnvName, validImage)
				e := testenv.CreateActiveEnvForTest()
				env := e.Start()

				sb := testenv.NewDiffingSandbox(subT, env)
				sb.StartManager(subT, controller.InitClusterReconcilerWithLogger(testLog))

				builder := testutil.NewBuilder("crdb").WithNodeCount(3).WithTLS().
					WithPVDataStore("1Gi").
					WithCockroachDBVersion(crdbVersion).WithImageObject(nil)

				require.NoError(subT, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(subT, sb, builder,
					e2e.CreateClusterTimeout)
				subT.Log("Done with basic invalid cluster with image other than crdb")
				require.NoError(subT, sb.Delete(builder.Cr()))

				e.Stop()
			},
		},
	}
	steps.Run(t)
}
