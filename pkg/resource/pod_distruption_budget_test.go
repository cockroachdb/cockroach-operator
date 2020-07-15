package resource_test

import (
	"testing"

	"fmt"

	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPDBBuilder(t *testing.T) {

	var MaxInt int32 = 1
	cluster := testutil.NewBuilder("test-cluster").Namespaced("test-ns").WithMaxUnavaible(&MaxInt)
	commonLabels := labels.Common(cluster.Cr())

	labelSelector, err := metav1.ParseToLabelSelector("app=" + cluster.Cr().Name)
	if err != nil {
		t.Errorf("unexpected error parsing label: %v", err)
	}

	max := intstr.FromInt(1)

	tests := []struct {
		name     string
		cluster  *resource.Cluster
		selector map[string]string
		expected *policy.PodDisruptionBudget
	}{
		{
			name:     "builds default discovery service",
			cluster:  cluster.Cluster(),
			selector: commonLabels.Selector(),
			expected: &policy.PodDisruptionBudget{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-cluster",
					Labels: map[string]string{},
				},
				Spec: policy.PodDisruptionBudgetSpec{
					MaxUnavailable: &max,
					Selector:       labelSelector,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := &policy.PodDisruptionBudget{}

			err := resource.PdbBuilder{
				Cluster:  tt.cluster,
				Selector: tt.selector,
			}.Build(actual)
			require.NoError(t, err)

			diff := cmp.Diff(tt.expected, actual, testutil.RuntimeObjCmpOpts...)
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}
