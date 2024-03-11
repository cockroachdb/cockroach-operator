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
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"sigs.k8s.io/kubetest2/pkg/process"
)

var (
	dir     string
	version string
)

func main() {
	flag.StringVar(&dir, "dir", ".", "the directory run in")
	flag.StringVar(&version, "version", "", "the new version to release")
	flag.Parse()

	steps := []Step{
		ValidateVersion(),
		EnsureUniqueVersion(func(cmd *exec.Cmd) error { return cmd.Run() }),
		CreateReleaseBranch(process.ExecJUnit),
		UpdateVersion(),
		UpdateChangelog(os.ReadFile),
		GenerateFiles(process.ExecJUnit),
	}

	if err := os.Chdir(dir); err != nil {
		bail(err)
	}

	for _, step := range steps {
		if err := step.Apply(version); err != nil {
			bail(err)
		}
	}
}

func bail(err error) {
	fmt.Fprintf(os.Stderr, "OOPS! An error occurred: %s\n", err)
	os.Exit(1)
}
