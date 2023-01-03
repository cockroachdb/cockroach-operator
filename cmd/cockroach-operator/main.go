/*
Copyright 2023 The Cockroach Authors

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
	"os"
	"strings"

	crdbv1alpha1 "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/controller"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	certDir                 = "/tmp/k8s-webhook-server/serving-certs"
	defaultLeaderElectionID = "crdb-operator.cockroachlabs.com"
	watchNamespaceEnvVar    = "WATCH_NAMESPACE"
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
	var metricsAddr, featureGatesString, leaderElectionID string
	var enableLeaderElection, skipWebhookConfig bool

	// use zap logging cli options
	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&featureGatesString, "feature-gates", "", "Feature gate to enable, format is a command separated list enabling features, for instance RunAsNonRoot=false")
	flag.StringVar(&leaderElectionID, "leader-election-id", defaultLeaderElectionID, "The ID to use for leader election")
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

	// Namespaces that are managed. Can be one of:
	//   (empty) - Watch all namespaces
	//   ns1 - Watch one watchNamespace
	//   ns1,ns2 - Watch multiple namespaces
	// See:
	// https://sdk.operatorframework.io/docs/building-operators/golang/operator-scope/#configuring-watch-namespaces-dynamically
	watchNamespace := os.Getenv(watchNamespaceEnvVar)

	mgrOpts := ctrl.Options{
		Scheme:             scheme,
		Namespace:          watchNamespace,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   leaderElectionID,
		Port:               9443,
		CertDir:            certDir,
	}

	if strings.Contains(watchNamespace, ",") {
		setupLog.Info("manager set up with multiple namespaces", "namespaces", watchNamespace)
		mgrOpts.Namespace = ""
		mgrOpts.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(watchNamespace, ","))
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), mgrOpts)
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
		if err := SetupWebhookTLS(ctx, os.Getenv("NAMESPACE"), certDir); err != nil {
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
