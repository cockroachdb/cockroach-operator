package resource

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type PublicServiceBuilder struct {
	*Cluster

	Selector map[string]string
}

func (b PublicServiceBuilder) Build() (runtime.Object, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   b.PublicServiceName(),
			Labels: map[string]string{},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Name: "grpc", Port: *b.Cluster.Spec().GRPCPort},
				{Name: "http", Port: *b.Cluster.Spec().HTTPPort},
			},
			Selector: b.Selector,
		},
	}

	return service, nil
}

func (b PublicServiceBuilder) Placeholder() runtime.Object {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.PublicServiceName(),
		},
	}
}
