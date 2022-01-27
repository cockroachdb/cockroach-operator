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

package exec

import (
	"os"

	"sigs.k8s.io/kubetest2/pkg/process"
)

// StartGKEKubeTest2 starts a GKE server.
func StartGKEKubeTest2(clusterName string, zone string, project string) error {
	args := []string{
		"gke",
		"--up",
		"--cluster-name",
		clusterName,
		"--version",
		"latest",
		"--zone",
		zone,
		"--project",
		project,
		"--ignore-gcp-ssh-key",
	}

	println("Up(): startin gke cluster...\n")
	// we want to see the output so use process.ExecJUnit
	return process.ExecJUnit("kubetest2", args, os.Environ())
}

// StopGKEKubeTest2 stops a GKE server.
func StopGKEKubeTest2(clusterName string, zone string, project string) error {
	args := []string{
		"gke",
		"--down",
		"--cluster-name",
		clusterName,
		"--version",
		"latest",
		"--zone",
		zone,
		"--project",
		project,
		"--ignore-gcp-ssh-key",
	}

	println("Down(): stopin gke cluster...\n")
	// we want to see the output so use process.ExecJUnit
	return process.ExecJUnit("kubetest2", args, os.Environ())
}

// GetGKEKubeconfig gets kubeconfig from kind
func GetGKEKubeConfig(clusterName string, zone string, project string) error {
	args := []string{
		"gcloud",
		"container",
		"clusters",
		"get-credentials",
		clusterName,
		"--zone",
		zone,
		"--project",
		project,
	}
	println("getting kubeconfig for cluster ...\n")
	// we want to see the output so use process.ExecJUnit
	return process.ExecJUnit("kind", args, os.Environ())
}
