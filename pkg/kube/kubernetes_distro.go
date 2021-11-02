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

package kube

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubernetesDistribution interface {
	Get(ctx context.Context, clientset kubernetes.Interface, log logr.Logger) (string, error)
}

type kubernetesDistribution struct{}

func NewKubernetesDistribution() KubernetesDistribution {
	return &kubernetesDistribution{}
}

// Get the Kubernetes Distribution Type
func (kd kubernetesDistribution) Get(ctx context.Context, clientset kubernetes.Interface, log logr.Logger) (string, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		msg := "cannot get nodes"
		log.Error(err, msg)
		return "", errors.Wrap(err, msg)
	} else if len(nodes.Items) == 0 {
		msg := "nodes length is zero"
		log.Error(err, msg)
		return "", errors.Wrap(err, msg)
	}

	nodeName := nodes.Items[0].Name

	// Get node object
	node, err := clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		msg := "cannot get node"
		log.Error(err, msg)
		return "", errors.Wrap(err, msg)
	}

	// You can read Kubernetes version from either KubeletVersion or KubeProxyVersion
	kubeletVersion := node.Status.NodeInfo.KubeletVersion

	if strings.Contains(kubeletVersion, "gke") {
		return "kubernetes-operator-gke", nil
	} else if strings.Contains(kubeletVersion, "aks") {
		return "kubernetes-operator-aks", nil
	} else if strings.Contains(kubeletVersion, "eks") {
		return "kubernetes-operator-eks", nil
	} else {
		for key := range node.Annotations {
			if strings.Contains("openshift", key) {
				return "kubernetes-operator-openshift", nil
			}
		}
	}

	log.V(int(zapcore.WarnLevel)).Info("found unknown kubernetes distribution")
	return "kubernetes-operator-unknown", nil
}

type mockKubernetesDistribution struct{}

func MockKubernetesDistribution() KubernetesDistribution {
	return &mockKubernetesDistribution{}
}

// Get the Kubernetes Distribution Type
func (mock mockKubernetesDistribution) Get(_ context.Context, _ kubernetes.Interface, _ logr.Logger) (string, error) {
	return "GKE", nil
}
