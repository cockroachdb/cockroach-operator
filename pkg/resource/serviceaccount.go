package resource

import (
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceAccountBuilder struct {
	*Cluster

	Selector map[string]string
}

func (b ServiceAccountBuilder) Build(obj runtime.Object) error {
	account, ok := obj.(*corev1.ServiceAccount)
	if !ok {
		return errors.New("obj was not a *corev1.ServiceAccount")
	}

	if account.ObjectMeta.Name == "" {
		account.ObjectMeta.Name = b.ServiceAccountName()
	}

	if account.ObjectMeta.Labels == nil {
		account.ObjectMeta.Labels = map[string]string{}
	}

	return nil
}

func (b ServiceAccountBuilder) Placeholder() runtime.Object {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.ServiceAccountName(),
		},
	}
}
