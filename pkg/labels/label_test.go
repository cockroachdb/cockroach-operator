package labels_test

import (
	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestDefaultCommonLabels(t *testing.T) {
	clusterAsPartOfApp := testutil.NewBuilder("test-cluster").Namespaced("test-ns").Cr()

	expected := map[string]string{
		"app.kubernetes.io/name":       "cockroachdb",
		"app.kubernetes.io/instance":   "test-cluster",
		"app.kubernetes.io/component":  "database",
		"app.kubernetes.io/part-of":    "cockroachdb",
		"app.kubernetes.io/managed-by": "cockroach-operator",
	}

	actual := labels.Common(clusterAsPartOfApp).AsMap()

	assert.Equal(t, expected, actual)
}

func TestPartOfAndVersionGetCustomized(t *testing.T) {
	clusterAsPartOfApp := testutil.NewBuilder("test-cluster").Namespaced("test-ns").Cr()
	clusterAsPartOfApp.Labels[labels.PartOfKey] = "django"
	clusterAsPartOfApp.Status.Version = "v19.2"

	expected := map[string]string{
		"app.kubernetes.io/name":       "cockroachdb",
		"app.kubernetes.io/instance":   "test-cluster",
		"app.kubernetes.io/version":    "v19.2",
		"app.kubernetes.io/component":  "database",
		"app.kubernetes.io/part-of":    "django",
		"app.kubernetes.io/managed-by": "cockroach-operator",
	}

	actual := labels.Common(clusterAsPartOfApp).AsMap()

	assert.Equal(t, expected, actual)
}

func TestFromObject(t *testing.T) {
	expected := map[string]string{
		"label1": "value1",
		"label2": "value2",
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: expected,
		},
	}

	actual, _ := labels.FromObject(service)

	assert.Equal(t, expected, actual.AsMap())
}

func TestUpdateRemovesManagedLabel(t *testing.T) {
	objLabels := map[string]string{
		"label1":                    "value1",
		"app.kubernetes.io/version": "v1.0",
	}

	update := map[string]string{
		"app.kubernetes.io/name": "cockroachdb",
	}

	expected := map[string]string{
		"label1":                 "value1",
		"app.kubernetes.io/name": "cockroachdb",
	}

	labels.Update(objLabels, update)

	assert.Equal(t, expected, objLabels)
}

func TestSelector(t *testing.T) {
	cr := testutil.NewBuilder("test-cluster").Namespaced("test-ns").Cr()

	expected := map[string]string{
		"app.kubernetes.io/name":      "cockroachdb",
		"app.kubernetes.io/instance":  "test-cluster",
		"app.kubernetes.io/component": "database",
	}

	assert.Equal(t, expected, labels.Common(cr).Selector())
}
