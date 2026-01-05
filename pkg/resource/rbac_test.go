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

package resource_test

import (
	"fmt"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestServiceAccountBuilder(t *testing.T) {
	cluster := testutil.
		NewBuilder("test-cluster").
		Namespaced("test-ns").
		WithAnnotations(map[string]string{"key": "test-discovery-svc"}).
		Cluster()

	expected := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.ServiceAccountName(),
			Namespace: cluster.Namespace(),
		},
	}

	builder := ServiceAccountBuilder{Cluster: cluster}
	created := builder.Placeholder()
	require.NoError(t, builder.Build(created))

	diff := cmp.Diff(expected, created, testutil.RuntimeObjCmpOpts...)
	if diff != "" {
		require.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
	}
}

func TestRoleBuilder(t *testing.T) {
	cluster := testutil.
		NewBuilder("test-cluster").
		Namespaced("test-ns").
		WithAnnotations(map[string]string{"key": "test-discovery-svc"}).
		Cluster()

	expected := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster-role",
			Namespace: cluster.Namespace(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"use"},
				APIGroups: []string{"security.openshift.io"},
				Resources: []string{"securitycontextconstraints"},
			},
		},
	}

	builder := RoleBuilder{Cluster: cluster}
	created := builder.Placeholder()
	require.NoError(t, builder.Build(created))

	diff := cmp.Diff(expected, created, testutil.RuntimeObjCmpOpts...)
	if diff != "" {
		require.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
	}
}

func TestRoleBindingBuilder(t *testing.T) {
	cluster := testutil.
		NewBuilder("test-cluster").
		Namespaced("test-ns").
		WithAnnotations(map[string]string{"key": "test-discovery-svc"}).
		Cluster()

	builder := RoleBindingBuilder{Cluster: cluster, ServiceAccountName: "test-cluster-sa"}

	withSubject := func(name string) *rbacv1.RoleBinding {
		bound := builder.Placeholder().(*rbacv1.RoleBinding)
		bound.Subjects = []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: name},
		}

		return bound
	}

	tests := []struct {
		name     string
		given    client.Object
		expected *rbacv1.RoleBinding
	}{
		{
			name:  "first subject",
			given: builder.Placeholder(),
			expected: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cluster.RoleBindingName(),
					Namespace: cluster.Namespace(),
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     cluster.RoleName(),
				},
				Subjects: []rbacv1.Subject{
					{Kind: "ServiceAccount", Name: cluster.ServiceAccountName()},
				},
			},
		},
		{
			name:  "second subject",
			given: withSubject("other-cluster-sa"),
			expected: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cluster.RoleBindingName(),
					Namespace: cluster.Namespace(),
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     cluster.RoleName(),
				},
				Subjects: []rbacv1.Subject{
					{Kind: "ServiceAccount", Name: "other-cluster-sa"},
					{Kind: "ServiceAccount", Name: cluster.ServiceAccountName()},
				},
			},
		},
		{
			name:  "subject already exists",
			given: withSubject("test-cluster-sa"),
			expected: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cluster.RoleBindingName(),
					Namespace: cluster.Namespace(),
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     cluster.RoleName(),
				},
				Subjects: []rbacv1.Subject{
					{Kind: "ServiceAccount", Name: cluster.ServiceAccountName()},
				},
			},
		},
	}

	for _, tt := range tests {
		require.NoError(t, builder.Build(tt.given), tt.name)

		diff := cmp.Diff(tt.expected, tt.given, testutil.RuntimeObjCmpOpts...)
		if diff != "" {
			require.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff), tt.name)
		}
	}
}
