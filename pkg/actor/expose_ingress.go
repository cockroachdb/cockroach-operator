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

package actor

import (
	"context"
	"fmt"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/util"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newExposeIngress(scheme *runtime.Scheme, cl client.Client, config *rest.Config, clientset kubernetes.Interface) Actor {
	return &exposeIngress{
		action:    newAction(scheme, cl, nil, clientset),
		config:    config,
		v1Ingress: util.CheckIfAPIVersionKindAvailable(config, "networking.k8s.io/v1", "Ingress"),
	}
}

// exposeIngress initializes and reconciles the ingress resource needed for exposing the CockroachDB cluster service
type exposeIngress struct {
	action
	config    *rest.Config
	v1Ingress bool
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

	ui := resource.UIIngressBuilder{Cluster: cluster, Labels: labelSelector, V1Ingress: ei.v1Ingress}
	sql := resource.SQLIngressBuilder{Cluster: cluster, Labels: labelSelector, V1Ingress: ei.v1Ingress}
	uiIngressEnabled := cluster.IsUIIngressEnabled()
	sqlIngressEnabled := cluster.IsSQLIngressEnabled()

	if cluster.IsIngressNeeded() {
		var (
			builders        []resource.Builder
			conditionsToSet []api.ClusterConditionType
		)

		if uiIngressEnabled {
			builders = append(builders, ui)
			conditionsToSet = append(conditionsToSet, api.CrdbUIIngressExposedCondition)
		}

		if sqlIngressEnabled {
			builders = append(builders, sql)
			conditionsToSet = append(conditionsToSet, api.CrdbSQLIngressExposedCondition)
		}

		for i, b := range builders {
			_, err := resource.Reconciler{
				ManagedResource: r,
				Builder:         b,
				Owner:           owner,
				Scheme:          ei.scheme,
			}.Reconcile()
			if err != nil {
				return errors.Wrapf(err, "failed to reconcile %s", b.ResourceName())
			}
			cluster.SetTrue(conditionsToSet[i])
		}

		if err := ei.client.Status().Update(ctx, cluster.Unwrap()); err != nil {
			msg := "failed to update IngressExposed condition in status"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}
	}

	uiIngressConditionTrue := condition.True(api.CrdbUIIngressExposedCondition, cluster.Status().Conditions)
	sqlIngressConditionTrue := condition.True(api.CrdbSQLIngressExposedCondition, cluster.Status().Conditions)

	// if ingress not needed but expose ingress condition is true, then its case of ingress update.
	// delete the ingress resource
	if !cluster.IsIngressNeeded() || (!uiIngressEnabled && uiIngressConditionTrue) || (!sqlIngressEnabled && sqlIngressConditionTrue) {
		var (
			builders        []resource.Builder
			conditionsToSet []api.ClusterConditionType
		)

		if !uiIngressEnabled && uiIngressConditionTrue {
			builders = append(builders, ui)
			conditionsToSet = append(conditionsToSet, api.CrdbUIIngressExposedCondition)
		}

		if !sqlIngressEnabled && sqlIngressConditionTrue {
			builders = append(builders, sql)
			conditionsToSet = append(conditionsToSet, api.CrdbSQLIngressExposedCondition)
		}

		for i, b := range builders {
			ing := b.Placeholder()
			ing.SetNamespace(cluster.Namespace())
			if err := ei.client.Delete(ctx, ing); err != nil {
				msg := fmt.Sprintf("failed to delete [%s] ingress resource", ing.GetName())
				log.Error(err, msg)
				return errors.Wrap(err, msg)
			}
			cluster.SetFalse(conditionsToSet[i])
		}

		if err := ei.client.Status().Update(ctx, cluster.Unwrap()); err != nil {
			msg := "failed to update IngressExposed condition in status"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}
	}

	log.Info("reconciled ingress resource")
	return nil
}
