/*
Copyright 2024 The Cockroach Authors

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
	"errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PublicServiceBuilder struct {
	*Cluster

	Selector map[string]string
}

func (b PublicServiceBuilder) ResourceName() string {
	return b.PublicServiceName()
}

func (b PublicServiceBuilder) Build(obj client.Object) error {
	service, ok := obj.(*corev1.Service)
	if !ok {
		return errors.New("failed to cast to Service object")
	}

	if service.ObjectMeta.Name == "" {
		service.ObjectMeta.Name = b.PublicServiceName()
	}

	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = map[string]string{}
	}

	service.Annotations = b.Spec().AdditionalAnnotations

	if service.Spec.Type != corev1.ServiceTypeClusterIP {
		service.Spec = corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Name: "grpc", Port: b.Cluster.GetGRPCPort()},
				{Name: "http", Port: b.Cluster.GetHTTPPort()},
				{Name: "sql", Port: b.Cluster.GetSQLPort()},
			},
		}
	}
	service.Spec.Selector = b.Selector

	return nil
}

func (b PublicServiceBuilder) Placeholder() client.Object {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.PublicServiceName(),
		},
	}
}
