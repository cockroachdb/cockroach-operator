/*
Copyright 2021 The Cockroach Authors

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

	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IngressBuilder models the Ingress that the operator maintains.
type IngressBuilder struct {
	*Cluster

	Labels map[string]string
}

func (b IngressBuilder) ResourceName() string {
	return b.IngressName()
}

func (b IngressBuilder) Build(obj client.Object) error {
	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		return errors.New("failed to cast to Ingress object")
	}

	if ingress.ObjectMeta.Name == "" {
		ingress.ObjectMeta.Name = b.ResourceName()
	}

	if ingress.ObjectMeta.Labels == nil {
		ingress.ObjectMeta.Labels = map[string]string{}
	}

	ingress.Labels = b.Labels
	ingress.Annotations = b.Spec().AdditionalAnnotations

	if ingress.ObjectMeta.Annotations == nil {
		ingress.ObjectMeta.Annotations = map[string]string{}
	}

	ingressConfig := b.Spec().Ingress

	kube.MergeAnnotations(ingress.ObjectMeta.Annotations, ingressConfig.Annotations)

	ingress.Spec = v1beta1.IngressSpec{
		IngressClassName: &ingressConfig.IngressClassName,
		TLS:              ingressConfig.TLS,
	}

	var rules []v1beta1.IngressRule

	if ingressConfig.HTTP != nil {
		rules = append(rules, getIngressRule(ingressConfig.HTTP.Host, b.PublicServiceName(), intstr.FromString("http")))
	}

	if ingressConfig.GRPC != nil {
		rules = append(rules, getIngressRule(ingressConfig.GRPC.Host, b.PublicServiceName(), intstr.FromString("grpc")))
	}

	if ingressConfig.SQL != nil {
		rules = append(rules, getIngressRule(ingressConfig.SQL.Host, b.PublicServiceName(), intstr.FromString("sql")))
	}

	ingress.Spec.Rules = rules

	return nil
}

func (b IngressBuilder) Placeholder() client.Object {
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.ResourceName(),
		},
	}
}

func getIngressRule(host, serviceName string, servicePort intstr.IntOrString) v1beta1.IngressRule {
	return v1beta1.IngressRule{
		Host: host,
		IngressRuleValue: v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath{
					{
						Backend: v1beta1.IngressBackend{
							ServiceName: serviceName,
							ServicePort: servicePort,
						},
					},
				},
			},
		},
	}
}
