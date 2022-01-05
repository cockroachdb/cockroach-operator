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

package env

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExpandPath expands a relative path to a workspace specific one. When running with bazel, this will be relative to the
// test workspace's build directory. Otherwise, it'll be relative to the project's root directory.
func ExpandPath(path ...string) string {
	if v, ok := os.LookupEnv("RUNFILES_DIR"); ok {
		prefix := []string{v, os.Getenv("TEST_WORKSPACE"), os.Getenv("BUILD_WORKSPACE_DIRECTORY")}
		return filepath.Join(append(prefix, path...)...)
	}

	// when not running with bazel
	return filepath.Join(append([]string{projectDir()}, path...)...)
}

// PrependToPath prepends the supplied path to PATH
func PrependToPath(path ...string) {
	os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Join(path...), os.Getenv("PATH")))
}

// projectDir returns the project root directory, determined by bazel. Panics if the directory can't be determined.
func projectDir() string {
	res, err := exec.Command("bazel", "info", "workspace").Output()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(res))
}
