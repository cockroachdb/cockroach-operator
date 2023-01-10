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

package actor_test

import (
	"context"
	"testing"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	. "github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func TestSetupRBACActionAct(t *testing.T) {
	ctx := context.Background()
	log := zapr.NewLogger(zaptest.NewLogger(t))
	scheme := testutil.InitScheme(t)

	cluster := testutil.
		NewBuilder("cockroachdb").
		Namespaced("bogus-ns").
		WithUID("cockroachdb-uid").
		WithPVDataStore("500Mu").
		WithNodeCount(1).
		Cluster()

	key := func(n string) types.NamespacedName {
		return types.NamespacedName{Name: n, Namespace: cluster.Namespace()}
	}

	t.Run("creates service account, role, and role-binding", func(t *testing.T) {
		client := testutil.NewFakeClient(scheme)
		config := &rest.Config{}
		actor := NewDirector(scheme, client, config, nil).GetActor(api.SetupRBACAction)
		require.NoError(t, actor.Act(ctx, cluster, log))

		sa := new(corev1.ServiceAccount)
		role := new(rbacv1.Role)
		binding := new(rbacv1.RoleBinding)

		require.NoError(t, client.Get(ctx, key(cluster.ServiceAccountName()), sa))
		require.NoError(t, client.Get(ctx, key(cluster.RoleName()), role))
		require.NoError(t, client.Get(ctx, key(cluster.RoleBindingName()), binding))

		require.Equal(t, sa.Namespace, role.Namespace)
		require.Equal(t, sa.Namespace, binding.Namespace)
		require.Contains(t, binding.Subjects, rbacv1.Subject{Kind: "ServiceAccount", Name: sa.Name})
	})
}
