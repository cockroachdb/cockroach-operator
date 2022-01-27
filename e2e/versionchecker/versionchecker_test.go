/*
Copyright 2022 The Cockroach Authors

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

package versionchecker_test

import (
	"encoding/json"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crdbv1alpha1 "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/e2e"
	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	testenv "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestVersionCheckerJobPodPending creates a cluster and then tries to run an unscheduleable version checker job.
// It checks that there is never more than one version checker job running.
func TestVersionCheckerJobPodPending(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	// Does not seem to like running in parallel
	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	builder := testutil.NewBuilder("crdb").Namespaced(sb.Namespace).WithNodeCount(3).WithTLS().
		WithImage(e2e.MajorVersion).
		WithPVDataStore("32Mi").
		WithResources(
			corev1.ResourceRequirements{
				// This is a hack to make the version checker pod unschedulable. There's likely a better way to do it.
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    apiresource.MustParse("1000"),
					corev1.ResourceMemory: apiresource.MustParse("1000T"),
				},
			})
	steps := testutil.Steps{
		{
			Name: "start an unschedulable job",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				testutil.RequireAtMostOneVersionCheckerJob(t, sb, 300*time.Second)
			},
		},
	}
	steps.Run(t)
}

func TestLoggingAPIValidCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	e := testenv.CreateActiveEnvForTest()
	env := e.Start()
	defer e.Stop()

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	logJson := []byte(`{"sinkss": "invalidvalue"}`)
	logConfig := make(map[string]interface{})
	require.NoError(t, json.Unmarshal(logJson, &logConfig))
	testutil.RequireLoggingConfigMap(t, sb, "logging-configmap", string(logJson))

	builder := testutil.NewBuilder("crdb").Namespaced(sb.Namespace).WithNodeCount(3).WithTLS().
		WithImage("cockroachdb/cockroach:v21.1.0").
		WithPVDataStore("32Mi").
		WithClusterLogging("logging-configmap")

	steps := testutil.Steps{
		{
			Name: "validate the logging API input",
			Test: func(t *testing.T) {
				require.NoError(t, sb.Create(builder.Cr()))
				crdbCluster := &crdbv1alpha1.CrdbCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      builder.Cr().Name,
						Namespace: builder.Cr().Namespace,
					},
				}
				// Added sleep to reconcile the CR.
				time.Sleep(2 * time.Second)
				require.NoError(t, sb.Get(crdbCluster))
				require.Equal(t, crdbCluster.Status.ClusterStatus, "Failed")
			},
		},
	}
	steps.Run(t)
}
