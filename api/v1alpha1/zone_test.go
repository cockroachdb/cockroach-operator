package v1alpha1_test

import (
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetName(t *testing.T) {
	cluster := testutil.NewBuilder("crdb").WithNodeCount(1).Cluster()

	tests := []struct {
		name string
		zone api.AvailabilityZone
		want string
	}{
		{
			name: "without prefix",
			zone: api.AvailabilityZone{
				StatefulSetSuffix: "",
			},
			want: "crdb",
		},
		{
			name: "with prefix",
			zone: api.AvailabilityZone{
				StatefulSetSuffix: "-a",
			},
			want: "crdb-a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.zone.Name(cluster.StatefulSetName()))
		})
	}
}
