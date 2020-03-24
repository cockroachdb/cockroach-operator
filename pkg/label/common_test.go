package label_test

import (
	"github.com/cockroachlabs/crdb-operator/pkg/label"
	optesting "github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPartOfIsLeftUntouchedIfCustomized(t *testing.T) {
	clusterAsPartOfApp := optesting.MakeStandAloneCluster(t)
	clusterAsPartOfApp.Labels[label.PartOfLabelKey] = "django"

	actual := label.MakeCommonLabels(clusterAsPartOfApp)

	checkLabel(t, actual, label.PartOfLabelKey, "django")
}

func checkLabel(t *testing.T, labels map[string]string, key string, expected string) {
	actual, ok := labels[key]
	require.True(t, ok, "the label key %s should be present", key)
	assert.Equal(t, expected, actual, "should keep the value the same")
}
