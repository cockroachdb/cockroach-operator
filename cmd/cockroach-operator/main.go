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
	"flag"
	"fmt"
	"os"

	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/cockroachdb/cockroach-operator/pkg/controller"

	crdbv1alpha1 "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
)

const WatchNamespaceEnvVar = "WATCH_NAMESPACE"

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = crdbv1alpha1.AddToScheme(scheme)
}

func main() {
	var metricsAddr, featureGatesString string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&featureGatesString, "feature-gates", "", "Feature gate to enable, format is a command separated list enabling features, for instance RunAsNonRoot=false")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	// If features gates are passed to the command line, use it (otherwise use featureGates from configuration)
	if featureGatesString != "" {
		if err := utilfeature.DefaultMutableFeatureGate.Set(featureGatesString); err != nil {
			setupLog.Error(err, "unable to parse feature-gates flag")
			os.Exit(1)
		}
	}

	namespace, err := GetWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get watch namespace")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Namespace:          namespace,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	reconciler := controller.InitClusterReconciler()
	if err = reconciler(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CrdbCluster")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// GetWatchNamespace returns the namespace the operation should be watching for changes
func GetWatchNamespace() (string, error) {
	ns, found := os.LookupEnv(WatchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", WatchNamespaceEnvVar)
	}
	return ns, nil
}
