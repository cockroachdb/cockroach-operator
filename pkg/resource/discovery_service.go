package resource

import (
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// This service only exists to create DNS entries for each pod in
// the StatefulSet such that they can resolve each other's IP addresses.
// It does not create a load-balanced ClusterIP and should not be used directly
// by clients in most circumstances.
type DiscoveryServiceBuilder struct {
	*Cluster

	Selector map[string]string
}

func (b DiscoveryServiceBuilder) Build(obj runtime.Object) error {
	service, ok := obj.(*corev1.Service)
	if !ok {
		return errors.New("failed to access Service object")
	}

	if service.ObjectMeta.Name == "" {
		service.ObjectMeta.Name = b.DiscoveryServiceName()
	}

	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = map[string]string{}
	}

	service.ObjectMeta.Annotations = b.monitoringAnnotations()

	service.Spec = corev1.ServiceSpec{
		ClusterIP:                "None",
		PublishNotReadyAddresses: true,
		Ports: []corev1.ServicePort{
			{Name: "grpc", Port: *b.Cluster.Spec().GRPCPort},
			{Name: "http", Port: *b.Cluster.Spec().HTTPPort},
		},
		Selector: b.Selector,
	}

	return nil
}

func (b DiscoveryServiceBuilder) Placeholder() runtime.Object {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.DiscoveryServiceName(),
		},
	}
}

func (b *DiscoveryServiceBuilder) monitoringAnnotations() map[string]string {
	return map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/path":   "_status/vars",
		"prometheus.io/port":   fmt.Sprint(*(b.Cluster.Spec().HTTPPort)),
	}
}
