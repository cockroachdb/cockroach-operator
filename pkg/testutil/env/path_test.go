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

package env_test

import (
	"os"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {
	env := map[string]string{
		"RUNFILES_DIR":              "a",
		"TEST_WORKSPACE":            "b",
		"BUILD_WORKSPACE_DIRECTORY": "c",
	}

	withEnv(env, func() {
		require.Equal(t, "a/b/c/pkg/test/env/path_test.go", ExpandPath("pkg/test/env/path_test.go"))
	})
}

func TestPrependToPath(t *testing.T) {
	withEnv(map[string]string{"PATH": "/usr/local/bin"}, func() {
		PrependToPath("hack", "bin")
		require.Equal(t, "hack/bin:/usr/local/bin", os.Getenv("PATH"))
	})
}

func withEnv(env map[string]string, fn func()) {
	toReset := make(map[string]string)
	toClear := make([]string, 0)

	for k, v := range env {
		if oldVal, ok := os.LookupEnv(k); ok {
			toReset[k] = oldVal
		} else {
			toClear = append(toClear, k)
		}

		if v == "__unset__" {
			os.Unsetenv(k)
			continue
		}

		os.Setenv(k, v)
	}

	defer func() {
		for k, v := range toReset {
			os.Setenv(k, v)
		}

		for _, k := range toClear {
			os.Unsetenv(k)
		}
	}()

	fn()
}
