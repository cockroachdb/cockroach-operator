package resource

import (
	"errors"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterRoleBuilder struct {
	*Cluster

	Selector map[string]string
}

func (b ClusterRoleBuilder) Build(obj runtime.Object) error {
	role, ok := obj.(*rbacv1.ClusterRole)
	if !ok {
		return errors.New("obj was not a *rbacv1.Cluster")
	}

	if role.ObjectMeta.Name == "" {
		role.ObjectMeta.Name = b.ClusterRoleName()
	}

	if role.ObjectMeta.Labels == nil {
		role.ObjectMeta.Labels = map[string]string{}
	}

	role.Rules = []rbacv1.PolicyRule{
		// Locality container
		{
			APIGroups: []string{""},
			Resources: []string{"nodes"},
			Verbs:     []string{"get"},
		},
		// cockroach-utils
		{
			APIGroups:     []string{""},
			ResourceNames: []string{b.StatefulSetName()},
			Resources:     []string{"statefulsets"},
			Verbs:         []string{"get"},
		},
	}

	return nil
}

func (b ClusterRoleBuilder) Placeholder() runtime.Object {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.RoleName(),
		},
	}
}
