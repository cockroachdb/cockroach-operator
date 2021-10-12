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
package resource_test

import (
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestIngressBuilder(t *testing.T) {
	annotations := map[string]string{"key": "test-ingress"}
	ingressClass := "test-ingress-class"
	cluster := testutil.NewBuilder("test-cluster").Namespaced("test-ns").WithAnnotations(annotations)
	commonLabels := labels.Common(cluster.Cr())
	selector := commonLabels.Selector(cluster.Cr().Spec.AdditionalLabels)

	tests := []struct {
		name     string
		cluster  *resource.Cluster
		expected *v1beta1.Ingress
	}{
		{
			name: "builds default discovery service",
			cluster: cluster.WithIngress(&v1alpha1.IngressConfig{
				IngressClassName: ingressClass,
				Annotations:      annotations,
				HTTP: &v1alpha1.IngressService{
					Enabled: true,
					Host:    "ui.test.com",
				},
				GRPC: &v1alpha1.IngressService{
					Enabled: true,
					Host:    "grpc.test.com",
				},
				SQL: &v1alpha1.IngressService{
					Enabled: true,
					Host:    "sql.test.com",
				},
			}).Cluster(),
			expected: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-cluster",
					Labels:      selector,
					Annotations: annotations,
				},
				Spec: v1beta1.IngressSpec{
					IngressClassName: &ingressClass,
					Rules: []v1beta1.IngressRule{
						{
							Host: "ui.test.com",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{Paths: []v1beta1.HTTPIngressPath{
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: "test-cluster-public",
											ServicePort: intstr.FromString("http"),
										},
									},
								}},
							},
						},
						{
							Host: "grpc.test.com",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{Paths: []v1beta1.HTTPIngressPath{
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: "test-cluster-public",
											ServicePort: intstr.FromString("grpc"),
										},
									},
								}},
							},
						},
						{
							Host: "sql.test.com",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{Paths: []v1beta1.HTTPIngressPath{
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: "test-cluster-public",
											ServicePort: intstr.FromString("sql"),
										},
									},
								}},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := &v1beta1.Ingress{}

			err := resource.IngressBuilder{
				Cluster: tt.cluster,
				Labels:  selector,
			}.Build(actual)
			require.NoError(t, err)

			diff := cmp.Diff(tt.expected, actual, testutil.RuntimeObjCmpOpts...)
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}
