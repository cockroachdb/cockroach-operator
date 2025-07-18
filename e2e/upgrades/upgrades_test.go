/*
Copyright 2025 The Cockroach Authors

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
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cockroachdb/cockroach-operator/e2e"
	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	testenv "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	resRequirements = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(e2e.DefaultCPULimit),
			corev1.ResourceMemory: resource.MustParse(e2e.DefaultMemoryLimit),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(e2e.DefaultCPURequest),
			corev1.ResourceMemory: resource.MustParse(e2e.DefaultMemoryRequest),
		},
	}
)

// TestUpgradesMinorVersion tests a minor version bump
func TestUpgradesMinorVersion(t *testing.T) {

	// We are testing a Minor Version Upgrade with
	// partition update
	// Going from v24.1.0 to v24.1.2

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
		WithImage(e2e.MinorVersion1).
		WithPVDataStore("1Gi").WithResources(resRequirements)

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "upgrades the cluster to the next patch version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Image.Name = e2e.MinorVersion2
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
				t.Log("Done with upgrade")
			},
		},
	}

	steps.Run(t)
}

// TestUpgradesMajorVersion24.1to24.2 tests a major version upgrade
func TestUpgradesMajorVersion24_1to24_2(t *testing.T) {

	// We are doing a major version upgrade here
	// 24.1.2 to 24.2.2

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
		WithImage(e2e.MinorVersion2).
		WithPVDataStore("1Gi").WithResources(resRequirements)

	steps := testutil.Steps{
		{
			Name: "creates a 1-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "upgrades the cluster to the next minor version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Image.Name = e2e.MajorVersion
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
				t.Log("Done with major upgrade")
			},
		},
	}

	steps.Run(t)
}

// TestUpgradesMajorVersion21_2To22_1 is another major version upgrade
func TestUpgradesMajorVersion21_2To22_1(t *testing.T) {

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
		WithImage("cockroachdb/cockroach:v21.2.16").
		WithPVDataStore("1Gi").WithResources(resRequirements)

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "upgrades the cluster to the next major version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Image.Name = "cockroachdb/cockroach:v22.1.10"
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))
				// we wait 10 min because we will be waiting 3 min for each pod because
				// v20.1.16 does not have curl installed
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
				t.Log("Done with major upgrade")
			},
		},
	}

	steps.Run(t)
}

func TestUpgradesMajorVersionSkippingInnovativeRelease24_3To25_1(t *testing.T) {

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
		WithImage("cockroachdb/cockroach:v24.3.4").
		WithPVDataStore("1Gi").WithResources(resRequirements)

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "upgrades the cluster to the next major version skipping innovative release",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Image.Name = "cockroachdb/cockroach:v25.1.0"
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))
				// we wait 10 min because we will be waiting 3 min for each pod.
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
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
	// Going from v24.1.0 to v24.1.2
	// Then rollback to v24.1.0

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").
		WithAutomountServiceAccountToken(true).
		WithTerminationGracePeriodSeconds(5).
		WithNodeCount(3).
		WithTLS().
		WithImage(e2e.MinorVersion1).
		WithPVDataStore("1Gi").
		WithResources(resRequirements)

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "upgrades the cluster to the next patch version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				updated := current.DeepCopy()
				updated.Spec.Image.Name = e2e.MinorVersion2
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
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
				updated.Spec.Image.Name = e2e.MinorVersion1
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
				testutil.RequireDbContainersToUseImage(t, sb, updated)
				t.Log("Done with downgrade")
			},
		},
	}

	steps.Run(t)
}

// TestUpgradeWithInvalidVersion tests the upgrade to non-existent version which will result in Failure.
func TestUpgradeWithInvalidVersion(t *testing.T) {
	// We are testing an upgrade with invalid version
	// Upgrade is going to fail.

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
		WithImage(e2e.MinorVersion1).
		WithPVDataStore("1Gi").
		WithResources(resRequirements)

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "upgrades the cluster to the next patch version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = e2e.NonExistentVersion
				require.NoError(t, sb.Update(current))

				testutil.RequireClusterInImagePullBackoff(t, sb, builder)
				testutil.RequireClusterInFailedState(t, sb, builder)
				t.Log("Upgrade failed with invalid image")
			},
		},
	}

	steps.Run(t)
}

// TestUpgradeWithInvalidImage tests the upgrade to the image which exists but not a valid image.
// Upgrade should fail in this case.
func TestUpgradeWithInvalidImage(t *testing.T) {
	// We are testing an upgrade with invalid image
	// Upgrade is going to fail.

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
		WithImage(e2e.MinorVersion1).
		WithPVDataStore("1Gi").
		WithResources(resRequirements)

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "upgrades the cluster to the next patch version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = e2e.InvalidImage
				require.NoError(t, sb.Update(current))

				testutil.RequireClusterInImagePullBackoff(t, sb, builder)
				testutil.RequireClusterInFailedState(t, sb, builder)
				t.Log("Upgrade failed with invalid image")
			},
		},
	}

	steps.Run(t)
}

// TestUpgradeWithMajorVersionExcludingMajorFeature test major version upgrade with skipping a major release.
// Upgrade should fail in this case as well
func TestUpgradeWithMajorVersionExcludingMajorFeature(t *testing.T) {
	// We are testing a major version Upgrade with skipping feature
	// Upgrade is going to fail due to non-support of skipping major versions.

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
		WithImage(e2e.SkipFeatureVersion).
		WithPVDataStore("1Gi").WithResources(resRequirements)

	steps := testutil.Steps{
		{
			Name: "creates a 1-node secure cluster",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "upgrades the cluster to the next minor version",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))

				current.Spec.Image.Name = e2e.MajorVersion
				require.NoError(t, sb.Update(current))

				testutil.RequireClusterInFailedState(t, sb, builder)
				t.Log("Done with major upgrade with skipping feature")
			},
		},
	}

	steps.Run(t)
}
