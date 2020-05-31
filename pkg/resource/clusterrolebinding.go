package resource

import (
	"errors"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterRoleBindingBuilder struct {
	*Cluster

	Selector map[string]string
}

func (b ClusterRoleBindingBuilder) Build(obj runtime.Object) error {
	crb, ok := obj.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return errors.New("obj was not a *rbacv1.ClusterRoleBinding")
	}

	if crb.ObjectMeta.Name == "" {
		crb.ObjectMeta.Name = b.ClusterRoleBindingName()
	}

	if crb.ObjectMeta.Labels == nil {
		crb.ObjectMeta.Labels = map[string]string{}
	}

	crb.RoleRef = rbacv1.RoleRef{
		Kind: "ClusterRole",
		Name: b.ClusterRoleName(),
	}

	crb.Subjects = []rbacv1.Subject{
		{
			Kind: "ServiceAccount",
			Name: b.ServiceAccountName(),
			// ???
			// Namespace: "default",
		},
	}

	return nil
}

func (b ClusterRoleBindingBuilder) Placeholder() runtime.Object {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.ClusterRoleBindingName(),
		},
	}
}
