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

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type clusterChecker struct {
	podLabels map[string][]string
}

func (cc *clusterChecker) checkPodsAreUp(k8s kubernetes.Interface) error {
	for name, labels := range cc.podLabels {
		labels = formatLabelSelectors(name, labels)

		for _, label := range labels {
			if err := retry(testPods(label, k8s)); err != nil {
				return err
			}

			klog.Info(label + " is running")
		}
	}

	return nil
}

func formatLabelSelectors(key string, values []string) []string {
	labels := make([]string, len(values))
	for i, value := range values {
		labels[i] = fmt.Sprintf("%s=%s", key, value)
	}

	return labels
}

// testPods looks up Pods via a label and does an exponetial backoff test
// that the Pods exist and are running.
func testPods(label string, clientset kubernetes.Interface) func() error {
	return func() error {
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
}
