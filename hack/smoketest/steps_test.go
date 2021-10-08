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

package main_test

import (
	"testing"
	"time"

	. "github.com/cockroachdb/cockroach-operator/hack/smoketest"
	"github.com/stretchr/testify/require"
)

type mockExecFn struct {
	cmd  string
	args []string
	env  []string
	err  error
}

func (m *mockExecFn) exec(cmd string, args, env []string) error {
	m.cmd = cmd
	m.args = args
	m.env = env
	return m.err
}

func TestStartKindCluster(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{name: "cluster1", version: "1.18.2"},
		{name: "cluster2", version: "1.22.1"},
	}

	for _, tt := range tests {
		fn := new(mockExecFn)
		require.NoError(t, StartKindCluster(tt.name, tt.version).Apply(fn.exec))

		expArgs := []string{
			"create",
			"cluster",
			"--name",
			tt.name,
			"--image",
			"kindest/node:v" + tt.version,
		}

		require.Equal(t, "kind", fn.cmd)
		require.Equal(t, expArgs, fn.args)
	}
}

func TestStopKindCluster(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, StopKindCluster("test-cluster").Apply(fn.exec))

	require.Equal(t, "kind", fn.cmd)
	require.Equal(t, []string{"delete", "cluster", "--name", "test-cluster"}, fn.args)
}

func TestApplyManifest(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, ApplyManifest("some/path.yaml").Apply(fn.exec))

	require.Equal(t, "kubectl", fn.cmd)
	require.Equal(t, []string{"apply", "-f", "some/path.yaml"}, fn.args)
}

func TestWaitForDeploymentAvailable(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, WaitForDeploymentAvailable("deploy-name").Apply(fn.exec))

	expArgs := []string{
		"wait",
		"--for",
		"condition=Available",
		"deploy/deploy-name",
		"--timeout",
		"2m",
	}

	require.Equal(t, "kubectl", fn.cmd)
	require.Equal(t, expArgs, fn.args)
}

func TestWaitForSecret(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, WaitForSecret("my-secret").Apply(fn.exec))

	require.Equal(t, "kubectl", fn.cmd)
	require.Equal(t, []string{"get", "secret", "my-secret"}, fn.args)
}

func TestWaitForStatefulSetRollout(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, WaitForStatefulSetRollout("my-sts").Apply(fn.exec))

	require.Equal(t, "kubectl", fn.cmd)
	require.Equal(t, []string{"rollout", "status", "-w", "sts/my-sts"}, fn.args)
}

func TestWaitForPodReady(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, WaitForPodReady("my-pod").Apply(fn.exec))

	expArgs := []string{
		"wait",
		"--for",
		"condition=Ready",
		"pod/my-pod",
		"--timeout",
		"2m",
	}

	require.Equal(t, "kubectl", fn.cmd)
	require.Equal(t, expArgs, fn.args)
}

func TestInitBankWorkload(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, InitBankWorkload().Apply(fn.exec))

	expArgs := []string{
		"exec",
		"-it",
		"cockroachdb-client-secure",
		"--",
		"cockroach",
		"workload",
		"init",
		"bank",
	}

	require.Equal(t, "kubectl", fn.cmd)
	require.Equal(t, expArgs, fn.args[:len(fn.args)-1])
}

func TestRunBankWorkload(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, RunBankWorkload(5*time.Second).Apply(fn.exec))

	expArgs := []string{
		"exec",
		"-it",
		"cockroachdb-client-secure",
		"--",
		"cockroach",
		"workload",
		"run",
		"bank",
		"--duration",
		"5s",
	}

	require.Equal(t, "kubectl", fn.cmd)
	require.Equal(t, expArgs, fn.args[:len(fn.args)-1])
}
