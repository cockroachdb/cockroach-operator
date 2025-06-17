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

package main_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/hack/release"
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

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		version string
		isErr   bool
	}{
		{version: "20.1.3"},
		{version: "20.1.3-beta.1", isErr: true},
		{version: "v20.1.3", isErr: true},
		{version: "20.1.3-beta", isErr: true},
		{version: "2.3.2-beta.A", isErr: true},
		{version: "2.3.2-beta.1A", isErr: true},
		{version: "v20.1.3-beta.1", isErr: true},
		{version: "20.1.3a", isErr: true},
		{version: "20.1.3-rc.1", isErr: true},
	}

	for _, tt := range tests {
		err := ValidateVersion().Apply(tt.version)
		if tt.isErr {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)
	}
}

func TestEnsureUniqueVersion(t *testing.T) {
	cmdFn := func(cmd *exec.Cmd) error {
		require.Equal(t, []string{"git", "tag"}, cmd.Args)

		_, err := io.WriteString(cmd.Stdout, "v1.7.0\nv1.7.1\nv2.1.0\n")
		return err
	}

	require.NoError(t, EnsureUniqueVersion(cmdFn).Apply("0.1.0"))
	require.Error(t, EnsureUniqueVersion(cmdFn).Apply("2.1.0"))

	t.Run("when executing command fails", func(t *testing.T) {
		cmdFn := func(cmd *exec.Cmd) error {
			_, _ = io.WriteString(cmd.Stderr, "command error")
			return fmt.Errorf("boom")
		}

		require.EqualError(
			t,
			EnsureUniqueVersion(cmdFn).Apply("2.1.0"),
			"failed to get tags: command error - boom",
		)
	})
}

func TestUpdateVersion(t *testing.T) {
	require.NoError(t, UpdateVersion().Apply("1.2.3"))

	v, err := os.ReadFile("version.txt")
	require.NoError(t, err)
	require.Equal(t, "1.2.3", string(v))

	info, err := os.Stat("version.txt")
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0644), info.Mode())
}

func TestCreateReleaseBranch(t *testing.T) {
	fn := new(mockExecFn)
	require.NoError(t, CreateReleaseBranch(fn.exec).Apply("1.2.3"))

	require.Equal(t, "git", fn.cmd)
	require.Equal(t, []string{"checkout", "-b", "release-1.2.3", "origin/master"}, fn.args)
	require.Equal(t, os.Environ(), fn.env)
}

func TestGenerateFiles(t *testing.T) {
	fn := new(mockExecFn)

	tests := []struct {
		version string
		args    []string
	}{
		{version: "2.1.0", args: []string{"release/gen-files", "CHANNELS=stable", "DEFAULT_CHANNEL=stable"}},
	}

	for _, tt := range tests {
		require.NoError(t, GenerateFiles(fn.exec).Apply(tt.version))
		require.Equal(t, "make", fn.cmd)
		require.Equal(t, tt.args, fn.args)
		require.Equal(t, os.Environ(), fn.env)
	}
}

func TestUpdateChangelog(t *testing.T) {
	input := `
# CHANGELOG yada yada yada
...
# [Unreleased](https://github.com/cockroachdb/cockroach-operator/compare/v1.0.0...master)

* Some unreleased content

# [v1.0.0](https://github.com/cockroachdb/cockroach-operator/compare/v0.9.0...v1.0.0)
...
# [v0.9.0](https://github.com/cockroachdb/cockroach-operator/compare/v0.8.0...v0.9.0)
`

	expected := `
# CHANGELOG yada yada yada
...
# [Unreleased](https://github.com/cockroachdb/cockroach-operator/compare/v1.1.0...master)

# [v1.1.0](https://github.com/cockroachdb/cockroach-operator/compare/v1.0.0...v1.1.0)

* Some unreleased content

# [v1.0.0](https://github.com/cockroachdb/cockroach-operator/compare/v0.9.0...v1.0.0)
...
# [v0.9.0](https://github.com/cockroachdb/cockroach-operator/compare/v0.8.0...v0.9.0)
`

	err := UpdateChangelog(func(_ string) ([]byte, error) { return []byte(input), nil }).Apply("1.1.0")
	require.NoError(t, err)

	data, err := os.ReadFile("CHANGELOG.md")
	require.NoError(t, err)
	require.Equal(t, string(data), expected)
}
