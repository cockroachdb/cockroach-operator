/*
Copyright 2026 The Cockroach Authors

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
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

//
// This go program verifies that specific pods are running
// in the kube-system namespace.   By doing this verification
// we are able to hopefully check that a Kubernetes cluster is up and running.
// Currently gke and k3d are supported and there labels for the pods differ.
// This program uses an exponetial backoff to all the creation of the kubernetes
// client and pod run checks to fail and then retry with the backoff.
//

var (
	// set to the correct checker type via -type flag
	provider string

	checkers = map[string]*clusterChecker{
		"gke": {
			podLabels: map[string][]string{
				"k8s-app": {
					"event-exporter",
					"fluentbit-gke",
					"gke-metrics-agent",
					"kube-dns",
					"kube-dns-autoscaler",
					"glbc",
					"metrics-server",
					"gcp-compute-persistent-disk-csi-driver",
				},
			},
		},
		"k3d": {
			podLabels: map[string][]string{
				"k8s-app": {
					"kube-dns",
					"metrics-server",
				},
			},
		},
	}
)

func main() {
	flag.StringVar(&provider, "type", "", "gke, or k3d")
	flag.Parse()

	checker, ok := checkers[provider]
	if !ok {
		panic("wrong '-type' flag value. 'gke', or 'k3d' values are supported")
	}

	klog.Infof("Checking that '%s' k8s cluster has started completely", provider)

	k8s, err := createClientSet()
	if err != nil {
		klog.Error(err)
		panic("unable to create kubeclient")
	}

	if err := checker.checkPodsAreUp(k8s); err != nil {
		klog.Error(err)
		panic("failed to find matching pods")
	}

	klog.Infof("%s k8s cluster is up and running", provider)
}

// createClientSet creates a core-kubernetes client
// using the default homedir kubeconfig.
// This func uses an exponetial backoff to create the client.
func createClientSet() (kubernetes.Interface, error) {
	var config *rest.Config
	var retErr error
	var clientset kubernetes.Interface

	// use the current context in kubeconfig
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// create the rest config with a backoff
	if err := retry(func() error {
		config, retErr = clientcmd.BuildConfigFromFlags("", kubeconfig)
		klog.V(8).Info("tried getting config")
		return retErr
	}); err != nil {
		klog.V(8).Info("config errored out")
		return nil, err
	}

	// create the client with a backoff
	if err := retry(func() error {
		clientset, retErr = kubernetes.NewForConfig(config)
		klog.V(8).Info("tried getting client")
		return retErr
	}); err != nil {
		klog.Error("getting client errored out")
		return nil, err
	}

	return clientset, nil
}

func retry(fn func() error) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 5 * time.Minute

	return backoff.Retry(fn, b)
}
