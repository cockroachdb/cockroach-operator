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

package deployer

import (
	"errors"
	"os"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/process"
)

// Down destroys a eks cluster
func (d *deployer) Down() error {
	if d.ClusterName == "" {
		return errors.New("flag cluster-name is not set")
	}

	file, err := getClusterFile(d.ClusterName)
	if err != nil {
		klog.Fatalf("unable to get yaml cluster file name %v", err)
		return err
	}

	args := []string{
		"eks", "delete", "cluster",
		"-p", file,
		"--enable-prompt=false",
	}

	// we want to see the output so use process.ExecJUnit
	// use the aws-k8s-tester binary to destroy the cluster
	if err := process.ExecJUnit("aws-k8s-tester", args, os.Environ()); err != nil {
		klog.Fatalf("unable to destroy eks cluster: %v", err)
		return err
	}

	return nil
}
