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
	"fmt"
	"os"
	"path/filepath"

	"errors"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/process"
)

// all of the tests that we disable that aws-k8s-tester will run
var disable = []string{
	"AWS_K8S_TESTER_EKS_ADD_ON_KUBERNETES_DASHBOARD_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_PROMETHEUS_GRAFANA_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_NLB_HELLO_WORLD_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_NLB_GUESTBOOK_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_ALB_2048_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_JOBS_PI_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_JOBS_ECHO_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_CRON_JOBS_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_CSRS_LOCAL_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_CONFIGMAPS_LOCAL_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_SECRETS_LOCAL_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_WORDPRESS_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_JUPYTER_HUB_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_CUDA_VECTOR_ADD_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_CLUSTER_LOADER_LOCAL_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_HOLLOW_NODES_LOCAL_ENABLE",
	"AWS_K8S_TESTER_EKS_ADD_ON_STRESSER_LOCAL_ENABLE",
}

// Up creates an EKS cluster.
func (d *deployer) Up() error {

	// disable the various tests
	for _, e := range disable {
		os.Setenv(e, "false")
	}

	// enable node groups so a ASG is created, and we have
	// worker nodes
	os.Setenv("AWS_K8S_TESTER_EKS_ADD_ON_NODE_GROUPS_ENABLE", "true")

	if d.ClusterName == "" {
		return errors.New("flag cluster-name is not set")
	}

	file, err := getClusterFile(d.ClusterName)
	if err != nil {
		klog.Fatalf("unable to get yaml cluster file name %v", err)
		return err
	}

	args := []string{
		"eks", "create", "cluster",
		"-p", file,
		"--enable-prompt=false",
	}

	// we want to see the output so use process.ExecJUnit
	if err := process.ExecJUnit("aws-k8s-tester", args, os.Environ()); err != nil {
		klog.Fatalf("unable to create eks cluster %v", err)
		return err
	}

	return nil
}

// IsUp is required by the kubetest2 interface, but the
// eks binary does this as part of the cluster creation.
func (d *deployer) IsUp() (up bool, err error) {
	return true, nil
}

func getClusterFile(clusterName string) (string, error) {
	if clusterName == "" {
		return "", errors.New("flag cluster-name is not set")
	}

	clusterYaml := fmt.Sprintf("%s-eks.yaml", clusterName)
	return filepath.Join(os.TempDir(), clusterYaml), nil
}
