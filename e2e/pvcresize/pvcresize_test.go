/*
Copyright 2024 The Cockroach Authors

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

package pvcresize

import (
	"context"
	"flag"
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
	apiresource "k8s.io/apimachinery/pkg/api/resource"
)

// run the pvc test
var pvc = flag.Bool("pvc", false, "run pvc test")

func TestPVCResize(t *testing.T) {
	// Testing PVCResize
	if !*pvc {
		t.Skip("platform does not support pvc resize")
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
		WithImage(e2e.MajorVersion).
		WithPVDataStore("1Gi")

	// This defaulting is done by webhook mutation config, but in tests we are doing it manually.
	builder.Cr().Default()

	steps := testutil.Steps{
		{
			Name: "creates a 3-node secure cluster db",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireClusterToBeReadyEventuallyTimeout(t, sb, builder, e2e.CreateClusterTimeout)
			},
		},
		{
			Name: "resize PVC",
			Test: func(t *testing.T) {
				current := builder.Cr()
				require.NoError(t, sb.Get(current))
				quantity := apiresource.MustParse("2Gi")

				updated := current.DeepCopy()
				updated.Spec.DataStore.VolumeClaim.PersistentVolumeClaimSpec.Resources.Requests[corev1.ResourceStorage] = quantity
				require.NoError(t, sb.Patch(updated, client.MergeFrom(current)))

				t.Log("updated CR")

				testutil.RequirePVCToResize(t, context.TODO(), sb, builder, quantity)
				t.Log("here resized")
			},
		},
	}
	steps.Run(t)
}
