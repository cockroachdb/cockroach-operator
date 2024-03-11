/*
Copyright 2024 The Cockroach Authors

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

package actor

import (
	"context"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newDeploy(scheme *runtime.Scheme, cl client.Client, kd kube.KubernetesDistribution, clientset kubernetes.Interface) Actor {
	return &deploy{
		action: newAction(scheme, cl, nil, clientset),
		kd:     kd,
	}
}

// deploy initializes and reconciles the Kubernetes resources needed by the CockroachDB cluster:
// services, a statefulset and a pod disruption budget
type deploy struct {
	action
	kd kube.KubernetesDistribution
}

// GetActionType returns the  api.DeployAction value used to set the cluster status errors
func (d deploy) GetActionType() api.ActionType {
	return api.DeployAction
}

func (d deploy) Act(ctx context.Context, cluster *resource.Cluster, log logr.Logger) error {
	log.V(DEBUGLEVEL).Info("reconciling resources on deploy action")

	owner := cluster.Unwrap()
	r := resource.NewManagedKubeResource(ctx, d.client, cluster, kube.AnnotatingPersister)

	kubernetesDistro, err := d.kd.Get(ctx, d.clientset, log)
	if err != nil {
		return errors.Wrap(err, "failed to get Kubernetes distribution")
	}

	labelSelector := r.Labels.Selector(cluster.Spec().AdditionalLabels)
	builders := []resource.Builder{
		resource.DiscoveryServiceBuilder{Cluster: cluster, Selector: labelSelector},
		resource.PublicServiceBuilder{Cluster: cluster, Selector: labelSelector},
		resource.StatefulSetBuilder{Cluster: cluster, Selector: labelSelector, Telemetry: kubernetesDistro},
		resource.PdbBuilder{Cluster: cluster, Selector: labelSelector},
	}

	for _, b := range builders {
		changed, err := resource.Reconciler{
			ManagedResource: r,
			Builder:         b,
			Owner:           owner,
			Scheme:          d.scheme,
		}.Reconcile()

		if err != nil {
			return errors.Wrapf(err, "failed to reconcile %s", b.ResourceName())
		}

		if changed {
			log.Info("created/updated a resource, stopping request processing", "resource", b.ResourceName())
			return nil
		}
	}

	log.Info("deployed database")
	return nil
}
