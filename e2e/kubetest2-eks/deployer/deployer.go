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

package deployer

import (
	"flag"
	"os"
	"path/filepath"
	"sync"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/pflag"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/types"
)

// Name is the name of the deployer
const Name = "eks"

// deployer is the base struct that
// defines flags for this deployer.
type deployer struct {
	// generic parts
	commonOptions types.Options
	// doInit helps to make sure the initialization is performed only once
	doInit sync.Once

	// Todo I am not certian we need KubeconfigPath

	KubeconfigPath string `flag:"kubeconfig" desc:"flag for the kubeconfig path"`

	ClusterName string `flag:"cluster-name" desc:"flag for the cluster name"`

	InstanceType string `flag:"instance-type" desc:"flag for the aws instance type name"`
	NumOfNodes   string `flag:"node-count" desc:"flag for the number of nodes"`
}

// assert that New implements types.NewDeployer
var _ types.NewDeployer = New

// assert that deployer implements types.Deployer
var _ types.DeployerWithKubeconfig = &deployer{}

// Provider returns the name of the deployer
func (d *deployer) Provider() string {
	return Name
}

// New implements deployer.New for EKS
func New(opts types.Options) (types.Deployer, *pflag.FlagSet) {

	// create a deployer object and set fields that are not flag controlled
	d := &deployer{
		commonOptions: opts,
	}

	// register flags
	fs := bindFlags(d)

	klog.InitFlags(nil)
	fs.AddGoFlagSet(flag.CommandLine)
	return d, fs
}

func bindFlags(d *deployer) *pflag.FlagSet {
	flags, err := gpflag.Parse(d)
	if err != nil {
		klog.Fatalf("unable to generate flags from deployer: %v", err)
		return nil
	}
	return flags
}

// Kubeconfig returns the path to the kubeconfig.
func (d *deployer) Kubeconfig() (string, error) {

	if d.KubeconfigPath != "" {
		return d.KubeconfigPath, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kube", "config"), nil
}
