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

package env_test

import (
	"os"
	"testing"

	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	. "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {
	env := map[string]string{
		"RUNFILES_DIR":              "a",
		"TEST_WORKSPACE":            "b",
		"BUILD_WORKSPACE_DIRECTORY": "c",
	}

	testutil.WithEnv(env, func() {
		require.Equal(t, "a/b/c/pkg/test/env/path_test.go", ExpandPath("pkg/test/env/path_test.go"))
	})
}

func TestPrependToPath(t *testing.T) {
	testutil.WithEnv(map[string]string{"PATH": "/usr/local/bin"}, func() {
		PrependToPath("hack", "bin")
		require.Equal(t, "hack/bin:/usr/local/bin", os.Getenv("PATH"))
	})
}
