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

package resource

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceAccountBuilder defines the database service account. This is the account that the crdb pods run as. Each instance of
// a CrdbCluster will have it's own service account.
type ServiceAccountBuilder struct {
	*Cluster
}

// ResourceName returns the name of the SA. This will be the instance name with a `-sa` suffix.
func (s ServiceAccountBuilder) ResourceName() string {
	return s.ServiceAccountName()
}

// Build populates the given object. For service accounts, there's nothing to do here. This function is added solely to
// prevent the Build function of the embedded struct from being called.
func (s ServiceAccountBuilder) Build(_ client.Object) error {
	return nil
}

// Placeholder defines the initial value of the ServiceAccount
func (s ServiceAccountBuilder) Placeholder() client.Object {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ResourceName(),
			Namespace: s.Namespace(),
		},
	}
}

// RoleBuilder defines the role that database service accounts belong to. This is a single role per namespace to which all
// CrdbCluster service accounts are bound.
type RoleBuilder struct {
	*Cluster
}

// ResourceName returns the name fo the Role. This is a constant value of "cockroach-role"
func (r RoleBuilder) ResourceName() string {
	return r.RoleName()
}

// Build populates the given object. For the role, there's nothing to do here. This function is added solely to
// prevent the Build function of the embedded struct from being called.
func (r RoleBuilder) Build(_ client.Object) error {
	return nil
}

// Placeholder defines the initial value for the Role. This role has limited rules defined, essentially only the
// ability to use security.openshift.io/securitycontextconstraints
func (r RoleBuilder) Placeholder() client.Object {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ResourceName(),
			Namespace: r.Namespace(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"use"},
				APIGroups: []string{"security.openshift.io"},
				Resources: []string{"securitycontextconstraints"},
			},
		},
	}
}

// RoleBindingBuilder defines the role binding for the namespace. It will bind `ServiceAccountName` to the Role.
type RoleBindingBuilder struct {
	*Cluster

	ServiceAccountName string
}

// ResourceName returns the name of the binding. This is a fixed value since there is only one per namespace.
func (r RoleBindingBuilder) ResourceName() string {
	return r.RoleBindingName()
}

// Build ensures that ServiceAccountName is bound to the Role. This operation is idempotent and safe to re-run.
func (r RoleBindingBuilder) Build(object client.Object) error {
	b := object.(*rbacv1.RoleBinding)

	if b.Subjects == nil {
		b.Subjects = []rbacv1.Subject{}
	}

	idx := -1
	for i, s := range b.Subjects {
		if s.Name == r.ServiceAccountName {
			idx = i
			break
		}
	}

	if idx < 0 {
		b.Subjects = append(b.Subjects, rbacv1.Subject{
			Kind: "ServiceAccount",
			Name: r.ServiceAccountName,
		})
	}

	return nil
}

// Placeholder defines the initial state for the binding. This sets up the role ref, but doesn't contain any subjects.
func (r RoleBindingBuilder) Placeholder() client.Object {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ResourceName(),
			Namespace: r.Namespace(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     r.RoleName(),
		},
	}
}
