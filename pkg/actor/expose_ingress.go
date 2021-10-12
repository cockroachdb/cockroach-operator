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

package actor

import (
	"context"
	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newExposeIngress(scheme *runtime.Scheme, cl client.Client, config *rest.Config, clientset kubernetes.Interface) Actor {
	return &exposeIngress{
		action: newAction(scheme, cl, nil, clientset),
		config: config,
	}
}

// exposeIngress initializes and reconciles the ingress resource needed for exposing the CockroachDB cluster service
type exposeIngress struct {
	action
	config *rest.Config
}

// GetActionType returns the  api.ExposeIngressAction value used to set the cluster status errors
func (ei exposeIngress) GetActionType() api.ActionType {
	return api.ExposeIngressAction
}

func (ei exposeIngress) Act(ctx context.Context, cluster *resource.Cluster, log logr.Logger) error {
	log.V(DEBUGLEVEL).Info("reconciling resource on expose ingress action")

	owner := cluster.Unwrap()
	r := resource.NewManagedKubeResource(ctx, ei.client, cluster, kube.AnnotatingPersister)

	labelSelector := r.Labels.Selector(cluster.Spec().AdditionalLabels)
	builder := resource.IngressBuilder{Cluster: cluster, Labels: labelSelector}

	if cluster.IsIngressNeeded() {
		_, err := resource.Reconciler{
			ManagedResource: r,
			Builder:         builder,
			Owner:           owner,
			Scheme:          ei.scheme,
		}.Reconcile()

		if err != nil {
			return errors.Wrapf(err, "failed to reconcile %s", builder.ResourceName())
		}

		cluster.SetTrue(api.CrdbIngressExposedCondition)

		if err = ei.client.Status().Update(ctx, cluster.Unwrap()); err != nil {
			msg := "failed to update IngressExposed condition in status"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}
		return nil
	}

	ingressConditionTrue := condition.True(api.CrdbIngressExposedCondition, cluster.Status().Conditions)

	// if ingress not needed but expose ingress condition is true, then its case of ingress update.
	// delete the ingress resource
	if !cluster.IsIngressNeeded() && ingressConditionTrue {
		ing := builder.Placeholder()
		ing.SetNamespace(cluster.Namespace())
		if err := ei.client.Delete(ctx, ing); err != nil {
			msg := "failed to delete the ingress resource"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}

		cluster.SetFalse(api.CrdbIngressExposedCondition)

		if err := ei.client.Status().Update(ctx, cluster.Unwrap()); err != nil {
			msg := "failed to update IngressExposed condition in status"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}
	}

	log.Info("reconciled ingress resource")
	return nil
}
