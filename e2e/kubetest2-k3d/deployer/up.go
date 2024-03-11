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

package deployer

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/kubetest2/pkg/exec"
	"sigs.k8s.io/kubetest2/pkg/metadata"
	"sigs.k8s.io/kubetest2/pkg/process"
)

func (d *deployer) IsUp() (bool, error) {
	// naively assume that if the api server reports nodes, the cluster is up
	lines, err := exec.CombinedOutputLines(
		exec.Command("kubectl", "get", "nodes", "-o=name"),
	)
	if err != nil {
		return false, metadata.NewJUnitError(err, strings.Join(lines, "\n"))
	}

	return len(lines) > 0, nil
}

func (d *deployer) Up() error {
	args := []string{
		"cluster",
		"create",
		d.ClusterName,
		"--wait",
	}

	if d.Servers > 0 {
		args = append(args, "--servers", fmt.Sprintf("%d", d.Servers))
	}

	if d.Image != "" {
		args = append(args, "--image", d.Image)
	}

	return process.ExecJUnit(
		"k3d",
		args,
		os.Environ(),
	)
}
