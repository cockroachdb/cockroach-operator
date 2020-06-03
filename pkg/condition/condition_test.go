package condition

import (
	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestInitializesForEmptyConditions(t *testing.T) {
	now := metav1.Now()

	status := api.CrdbClusterStatus{}

	expected := []api.ClusterCondition{
		{
			Type:               "NotInitialized",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: now,
		},
	}

	InitConditionsIfNeeded(&status, now)

	assert.ElementsMatch(t, expected, status.Conditions)
}
