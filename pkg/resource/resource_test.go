package resource_test

import (
	"context"
	"fmt"
	"github.com/cockroachlabs/crdb-operator/pkg/labels"
	"github.com/cockroachlabs/crdb-operator/pkg/ptr"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	amtypes "k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestReconcile(t *testing.T) {
	ctx := context.TODO()
	scheme := testutil.InitScheme(t)
	client := testutil.NewFakeClient(scheme)

	cluster := testutil.NewBuilder("test-cluster").Namespaced("default").Cluster()
	commonLabels := labels.Common(cluster.Unwrap())

	builder := resource.DiscoveryServiceBuilder{
		Cluster:  cluster,
		Selector: commonLabels.Selector(),
	}

	r := resource.Reconciler{
		ManagedResource: resource.NewManagedKubeResource(ctx, client, cluster),
		Builder:         builder,
		Owner:           cluster.Unwrap(),
		Scheme:          scheme,
	}

	upserted, err := r.Reconcile()
	require.NoError(t, err)

	assert.True(t, upserted)

	actual := &corev1.Service{}
	assert.NoError(t, client.Get(ctx, amtypes.NamespacedName{Name: "test-cluster", Namespace: "default"}, actual))

	expected := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
			Labels: map[string]string{
				"app.kubernetes.io/component":  "database",
				"app.kubernetes.io/instance":   "test-cluster",
				"app.kubernetes.io/managed-by": "crdb-operator",
				"app.kubernetes.io/name":       "cockroachdb",
				"app.kubernetes.io/part-of":    "cockroachdb",
			},
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/path":   "_status/vars",
				"prometheus.io/port":   "8080",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "crdb.cockroachlabs.com/v1alpha1",
					Kind:               "CrdbCluster",
					Name:               "test-cluster",
					Controller:         ptr.Bool(true),
					BlockOwnerDeletion: ptr.Bool(true),
				},
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
				"app.kubernetes.io/instance":  "test-cluster",
				"app.kubernetes.io/component": "database",
			},
		},
	}

	diff := cmp.Diff(expected, actual, testutil.RuntimeObjCmpOpts...)
	if diff != "" {
		assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
	}
}
