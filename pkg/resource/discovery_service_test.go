package resource_test

import (
	"context"
	"errors"
	"fmt"
	crdbv1alpha1 "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	optesting "github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestDiscoveryService_Reconcile(t *testing.T) {
	var customGrpcPort int32 = 2222

	ctx := context.TODO()
	cluster := optesting.MakeStandAloneCluster(t)
	updatedCluster := cluster.DeepCopy()
	updatedCluster.Spec.GrpcPort = &customGrpcPort

	expected, existing, updated := makeServiceFixtures(cluster.GetNamespace(), customGrpcPort)

	scheme := optesting.InitScheme(t)

	tests := []struct {
		name             string
		clusterSpec      *crdbv1alpha1.CrdbCluster
		preexistingObjs  []runtime.Object
		expected         *corev1.Service
		reactionInjector func(*optesting.FakeClient)
		wantErr          string
	}{
		{
			name:        "service definition can't be retrieved",
			clusterSpec: cluster,
			expected:    nil,
			wantErr:     "failed to fetch discovery service: test-ns/cockroachdb: internal error",
			reactionInjector: func(c *optesting.FakeClient) {
				c.AddReactor(
					"get",
					"services",
					func(action optesting.Action) (bool, error) {
						return true, errors.New("internal error")
					})
			},
		},
		{
			name:        "service resource does not exist",
			clusterSpec: cluster,
			expected:    expected,
		},
		{
			name:            "existing service spec matches the desired",
			clusterSpec:     cluster,
			preexistingObjs: []runtime.Object{existing},
			expected:        nil,
		},
		{
			name:        "updated cluster spec produces updated service spec",
			clusterSpec: updatedCluster,
			expected:    updated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := optesting.NewFakeClient(scheme, tt.preexistingObjs...)
			if tt.reactionInjector != nil {
				tt.reactionInjector(client)
			}

			actual, err := (&resource.DiscoveryService{Cluster: tt.clusterSpec}).Reconcile(ctx, client)

			if tt.wantErr != "" {
				assert.Nil(t, actual)
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err, "failed to reconcile due: %v", err)

			diff := cmp.Diff(tt.expected, actual, cmpopts.IgnoreTypes(metav1.TypeMeta{}))
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}

func makeServiceFixtures(ns string, customGrpcPort int32) (expected, existing, updated *corev1.Service) {
	expected = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb",
			Namespace: ns,
			Labels: map[string]string{
				"app.kubernetes.io/name":                                 "cockroachdb",
				"app.kubernetes.io/instance":                             "test-ns/test-cluster",
				"app.kubernetes.io/version":                              "v19.2",
				"app.kubernetes.io/component":                            "database",
				"app.kubernetes.io/part-of":                              "cockroachdb",
				"app.kubernetes.io/managed-by":                           "crdb-operator",
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/path":   "_status/vars",
				"prometheus.io/port":   "8080",
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:                "None",
			PublishNotReadyAddresses: true,
			Ports: []corev1.ServicePort{
				{Name: "grpc", Port: 26257},
				{Name: "http", Port: 8080},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name":      "cockroachdb",
				"app.kubernetes.io/instance":  "test-ns/test-cluster",
				"app.kubernetes.io/component": "database",
			},
		},
	}

	// Copy of expected with ports re-ordered to ensure the code under test isn't order-sensitive
	existing = expected.DeepCopy()
	existing.Spec.Ports[0], existing.Spec.Ports[1] = existing.Spec.Ports[1], existing.Spec.Ports[0]

	updated = expected.DeepCopy()
	for i, p := range updated.Spec.Ports {
		if p.Name == "grpc" {
			updated.Spec.Ports[i].Port = customGrpcPort
			break
		}
	}

	return
}
