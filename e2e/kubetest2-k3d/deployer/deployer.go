/*
Copyright 2025 The Cockroach Authors

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

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"sigs.k8s.io/kubetest2/pkg/types"
)

const (
	// Name is the name of the deployer
	Name = "k3d"

	version = "0.1.0-alpha"
)

var (
	// assert that New implements types.NewDeployer
	_ types.NewDeployer = New

	// ensure deployer satisfies types.DeployerWithKubeconfig
	_ types.DeployerWithKubeconfig = &deployer{}
)

// New implements Deployer.New for k3d
func New(opts types.Options) (types.Deployer, *pflag.FlagSet) {
	dep := &deployer{
		commonOptions: opts,
		logsDir:       filepath.Join(opts.RunDir(), "logs"),
	}

	return dep, dep.bindFlags()
}

type deployer struct {
	commonOptions types.Options
	logsDir       string

	ClusterName    string `flag:"cluster-name" desc:"the name of the cluster"`
	Image          string `flag:"image" desc:"the k3s image to use for the nodes"`
	KubeConfigPath string `flag:"kubeconfig" desc:"the kubeconfig to use"`
	Servers        int    `flag:"servers" desc:"the number of servers to run"`
}

func (d *deployer) bindFlags() *pflag.FlagSet {
	flags, err := gpflag.Parse(d)
	if err != nil {
		klog.Fatalf("unable to generate flags from deployer")
		return nil
	}

	klog.InitFlags(nil)
	flags.AddGoFlagSet(flag.CommandLine)

	return flags
}

func (d *deployer) Kubeconfig() (string, error) {
	if d.KubeConfigPath != "" {
		return d.KubeConfigPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".kube", "config"), nil
}

func (d *deployer) Version() string {
	return version
}
