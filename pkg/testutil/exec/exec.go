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

package exec

import (
	"os"

	"sigs.k8s.io/kubetest2/pkg/process"
)

// StopKubeTest2 stops a kind server.
func StopKubeTest2(clusterName string) error {
	args := []string{"kind", "--down", "--cluster-name",
		clusterName}

	println("Down(): stopin kind cluster...\n")
	// we want to see the output so use process.ExecJUnit
	return process.ExecJUnit("kubetest2", args, os.Environ())
}

// StartKubeTest2 starts a kind server.
func StartKubeTest2(clusterName string) error {
	args := []string{
		"kind", "--up", "--cluster-name", clusterName, "--loglevel", "10",
	}

	println("Up(): startin kind cluster...\n")
	// we want to see the output so use process.ExecJUnit
	return process.ExecJUnit("kubetest2", args, os.Environ())
}
