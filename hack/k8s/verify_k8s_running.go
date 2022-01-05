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

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
// Currently gke and kind are supported and there labels for the pods differ.
// This program uses an exponetial backoff to all the creation of the kubernetes
// client and pod run checks to fail and then retry with the backoff.
//

// backoffFactory is a replacable global for backoff creation. It may be
// replaced with shorter times to allow testing of Wait___ functions without
// waiting the entire default period
var backoffFactory = defaultBackoffFactory

func defaultBackoffFactory(maxTime time.Duration) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxTime
	return b
}

var defaultTime time.Duration = 5 * time.Minute

// K8sPodLabels are used to store information about
// kubernetes pod labels to look up pods
type K8sPodLabels struct {
	key    *string
	values *[]string
}

func createK8sPodLabel(key *string, values *[]string) *K8sPodLabels {
	return &K8sPodLabels{
		key:    key,
		values: values,
	}
}

// createLableValues takes a K8sPodLabels struct and
// returns a slice of values where the podLabels.key is
// appended to each of the strings in podLabels.values.
func createLabelValues(podLabels *K8sPodLabels) *[]string {
	values := *podLabels.values
	key := *podLabels.key
	labels := make([]string, len(values))

	for i := 0; i < len(values); i++ {
		labels[i] = fmt.Sprintf("%s=%s", key, values[i])
	}

	return &labels
}

// createClientSet creates a core-kubernetes client
// using the default homedir kubeconfig.
// This func uses an exponetial backoff to create the client.
func createClientSet() (*kubernetes.Clientset, error) {

	var config *rest.Config
	var retErr error
	var clientset *kubernetes.Clientset

	// use the current context in kubeconfig
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// create the rest config with a backoff
	err := backoff.Retry(func() error {
		config, retErr = clientcmd.BuildConfigFromFlags("", kubeconfig)
		klog.V(8).Info("tried getting config")
		return retErr
	}, backoffFactory(defaultTime))

	if err != nil {
		klog.V(8).Info("config errored out")
		return nil, err
	}

	// create the client with a backoff
	err = backoff.Retry(func() error {
		clientset, retErr = kubernetes.NewForConfig(config)
		klog.V(8).Info("tried getting client")
		return retErr
	}, backoffFactory(defaultTime))

	if err != nil {
		klog.Error("getting client errored out")
		return nil, err
	}

	return clientset, nil
}

// checkPodsAreUP interates through a slice of labels built from a K8sPodLabel
// struct and checks that the Pods are up for that label.
func checkPodsAreUp(podLabels *K8sPodLabels, clientset *kubernetes.Clientset) error {
	labels := createLabelValues(podLabels)

	for _, label := range *labels {
		err := testPods(label, clientset)
		if err != nil {
			return err
		}
		klog.Info(label + " is running")
	}

	return nil
}

// testPods looks up Pods via a label and does an exponetial backoff test
// that the Pods exist and are running.
func testPods(label string, clientset *kubernetes.Clientset) error {

	f := func() error {
		// find a list of pods via the label
		pods, err := clientset.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{
			LabelSelector: label,
		})
		if err != nil {
			klog.V(8).Info("getting pods errored out")
			return err
		}
		// if the pod list is zero no pods are running yet
		// and we throw an error.
		if len(pods.Items) == 0 {
			klog.V(8).Info("cannot find pods")
			return errors.New("unable to find any pods")
		}

		// iterate through the pods and test that each pod
		// is ready
		for _, pod := range pods.Items {
			if !kube.IsPodReady(&pod) {
				klog.V(8).Info("cannot find pods")
				return errors.New("pod not ready")
			}
		}

		return nil
	}

	b := backoffFactory(defaultTime)
	// run the func with a backoff factory
	return backoff.Retry(f, backoff.WithContext(b, context.TODO()))
}

//
// All of these values may change if kind or gke changes the pod labels and names
// that are running inside of kube-system.
//

// GKEPodLabels are the current gke
// pods that are running inside of kube-system.
// These pods are looked up via k8s-app=%s
var GKEPodLabels = []string{
	"event-exporter",
	"fluentbit-gke",
	"gke-metrics-agent",
	"kube-dns",
	"kube-dns-autoscaler",
	"glbc",
	"metrics-server",
	"gcp-compute-persistent-disk-csi-driver",
}

// K8sApp is the key that is used by pods
var K8sApp string = "k8s-app"

// Component is another key used by pods
var Component string = "component"

// KindPodLabels are the current kind
// pods that are running inside of kube-system.
// These pods are looked up via k8s-app=%s
var KindPodLabels = []string{"kindnet", "kube-proxy"}

// KindPodLabels are the current kind
// pods that are running inside of kube-system.
// These pods are looked up via component=%s
var KindComponentLabels = []string{"etcd", "kube-apiserver", "kube-controller-manager", "kube-scheduler"}

func main() {
	// -type is the flag we use to determine if we are checking kind or gke
	k8sTypeFlag := flag.String("type", "", "gke or kind")
	flag.Parse()

	if *k8sTypeFlag == "" {
		panic("set type arg")
	}

	klog.Infof("Checking that %s k8s cluster has started completely", *k8sTypeFlag)
	switch {
	case *k8sTypeFlag == "gke":
		p := createK8sPodLabel(&K8sApp, &GKEPodLabels)
		clientset, err := createClientSet()
		if err != nil {
			klog.Error(err)
			panic("unable to create kubeclient")
		}
		err = checkPodsAreUp(p, clientset)
		if err != nil {
			klog.Error(err)
			panic("unable to find gke pods")
		}
		// TODO(pseudomuto): This clearly isn't the right solution. Rather than sleeping and hoping for the best, we should
		// figure out what exactly is requiring this period of time and wait on that directly.
		time.Sleep(1 * time.Minute)
		fmt.Println("gke is running")
	case *k8sTypeFlag == "kind":
		p := createK8sPodLabel(&Component, &KindComponentLabels)
		clientset, err := createClientSet()
		if err != nil {
			klog.Error(err)
			panic("unable to create kubeclient")
		}
		err = checkPodsAreUp(p, clientset)
		if err != nil {
			klog.Error(err)
			panic("unable to find kind component pods")
		}

		p = createK8sPodLabel(&K8sApp, &KindPodLabels)
		err = checkPodsAreUp(p, clientset)
		if err != nil {
			klog.Error(err)
			panic("unable to find kind pods")
		}
	default:
		panic("wrong '-type' flag value. 'gke' and 'kind' values are supported")
	}

	klog.Infof("%s k8s cluster is up and running", *k8sTypeFlag)
}
