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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	testExec "github.com/cockroachdb/cockroach-operator/pkg/testutil/exec"
	"sigs.k8s.io/kubetest2/pkg/process"
)

func main() {
	os.Setenv("KUBECONFIG", "/tmp/.kube/config")
	os.Setenv("PATH", fmt.Sprintf("%s:%s", pathDirs(), os.Getenv("PATH")))
	os.Setenv("WATCH_NAMESPACE", "default")

	h := helper{
		cluster:      "test",
		workspaceDir: os.Getenv("BUILD_WORKSPACE_DIRECTORY"),
		env:          os.Environ(),
	}

	mayPanic(h.startKind)
	defer h.stopKind()

	mayPanic(h.applyOperator)
	mayPanic(h.run)
}

func pathDirs() string {
	// all args are expected to be paths to a bindary we need.
	paths := make([]string, len(os.Args)-1)
	for i, arg := range os.Args[1:] {
		paths[i] = filepath.Dir(arg)
	}

	return strings.Join(paths, ":")
}

func mayPanic(f func() error) {
	if err := f(); err != nil {
		panic(err)
	}
}

type helper struct {
	cluster      string
	workspaceDir string
	env          []string
}

func (h *helper) startKind() error {
	return testExec.StartKubeTest2(h.cluster)
}

func (h *helper) stopKind() error {
	return testExec.StopKubeTest2(h.cluster)
}

func (h *helper) applyOperator() error {
	return process.ExecJUnit(
		"kubectl",
		[]string{"apply", "-k", filepath.Join(h.workspaceDir, "config", "crd")},
		h.env,
	)
}

func (h *helper) run() error {
	return process.ExecJUnit(
		"cockroach-operator",
		[]string{},
		h.env,
	)
}

func (h *helper) getenv(key string) string {
	v := ""
	for _, pair := range h.env {
		if strings.HasPrefix(pair, key+"=") {
			parts := strings.SplitN(pair, "=", 2)
			v = parts[1]
		}
	}

	return v
}
