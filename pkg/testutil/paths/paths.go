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
package paths

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// MaybeSetEnv sets an environment variable that points to a binary. If that
// binary does not exist in the path this func throws a panic.
func MaybeSetEnv(key, bin string, path ...string) {
	if os.Getenv(key) != "" && key != "PATH" {
		return
	}
	p, err := getPath(bin, path...)
	if err != nil {
		panic(fmt.Sprintf(`Failed to find integration test dependency %q.
Either re-run this test using "bazel test //e2e/{name}" or set the %s environment variable.`, bin, key))
	}
	if key == "PATH" {
		p = p[:len(p)-len("/"+bin)]
		p = p + ":" + os.Getenv("PATH")
	}
	os.Setenv(key, p)
}

func getPath(name string, path ...string) (string, error) {
	bazelPath := filepath.Join(append([]string{os.Getenv("RUNFILES_DIR"), os.Getenv("TEST_WORKSPACE")}, path...)...)
	p, err := exec.LookPath(bazelPath)
	if err == nil {
		return p, nil
	}

	return exec.LookPath(name)
}
