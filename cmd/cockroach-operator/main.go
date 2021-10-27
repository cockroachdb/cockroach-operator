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
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	crdbv1alpha1 "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	certDir              = "/tmp/k8s-webhook-server/serving-certs"
	watchNamespaceEnvVar = "WATCH_NAMESPACE"
)

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
	var enableLeaderElection, skipWebhookConfig bool

	// use zap logging cli options
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&featureGatesString, "feature-gates", "", "Feature gate to enable, format is a command separated list enabling features, for instance RunAsNonRoot=false")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&skipWebhookConfig, "skip-webhook-config", false,
		"When set, don't setup webhook TLS certificates. Useful in OpenShift where this step is handled already.")
	flag.Parse()

	// create logger using zap cli options
	// for instance --zap-log-level=debug
	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)

	// If features gates are passed to the command line, use it (otherwise use featureGates from configuration)
	if featureGatesString != "" {
		if err := utilfeature.DefaultMutableFeatureGate.Set(featureGatesString); err != nil {
			setupLog.Error(err, "unable to parse feature-gates flag")
			os.Exit(1)
		}
	}

	namespace, err := getOperatorNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get watch namespace")
		os.Exit(1)
	}

	var options manager.Options
	if utilfeature.DefaultMutableFeatureGate.Enabled(features.MultipleNamespaces) {
		logger.Info("Watch and control all CrdbClusters in all namespaces.")
		options = ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: metricsAddr,
			LeaderElection:     enableLeaderElection,
			Port:               9443,
			CertDir:            certDir,
		}
	} else {
		logger.Info("Watch and control all CrdbClusters in single namespace. ", "namespace", namespace)
		options = ctrl.Options{
			Scheme:             scheme,
			Namespace:          namespace,
			MetricsBindAddress: metricsAddr,
			LeaderElection:     enableLeaderElection,
			Port:               9443,
			CertDir:            certDir,
		}
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	if err := (&crdbv1alpha1.CrdbCluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup webhook")
		os.Exit(1)
	}

	reconciler := controller.InitClusterReconciler()
	if err = reconciler(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CrdbCluster")
		os.Exit(1)
	}

	// add a logger to the main context
	ctx := logr.NewContext(ctrl.SetupSignalHandler(), logger)

	if !skipWebhookConfig {
		// ensure TLS is all set up for webhooks
		if err := SetupWebhookTLS(ctx, namespace, certDir); err != nil {
			setupLog.Error(err, "failed to setup TLS")
			os.Exit(1)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// getOperatorNamespace returns the namespace the operation should be watching for changes
func getOperatorNamespace() (string, error) {
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
