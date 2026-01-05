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

package actor

import (
	"context"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newSetupRBACAction(scheme *runtime.Scheme, cl client.Client) Actor {
	return &setupRBACAction{
		action: newAction(scheme, cl, nil, nil),
	}
}

type setupRBACAction struct {
	action
}

func (a *setupRBACAction) Act(ctx context.Context, cluster *resource.Cluster, log logr.Logger) error {
	log.V(DEBUGLEVEL).Info("creating service account, role, and role bindings")

	r := resource.NewManagedKubeResource(ctx, a.client, cluster, kube.AnnotatingPersister)
	owner := cluster.Unwrap()

	serviceAcct := resource.ServiceAccountBuilder{Cluster: cluster}
	builders := []resource.Builder{
		serviceAcct,
		resource.RoleBuilder{Cluster: cluster},
		resource.RoleBindingBuilder{Cluster: cluster, ServiceAccountName: serviceAcct.ResourceName()},
	}

	for _, b := range builders {
		_, err := resource.Reconciler{
			ManagedResource: r,
			Builder:         b,
			Owner:           owner,
			Scheme:          a.scheme,
		}.Reconcile()

		if err != nil {
			return errors.Wrapf(err, "failed to reconcile %s", b.ResourceName())
		}
	}

	log.Info("Finished setting up RBAC", "cluster", cluster.Name(), "namespace", cluster.Namespace())
	return nil
}

func (a *setupRBACAction) GetActionType() api.ActionType {
	return api.SetupRBACAction
}
