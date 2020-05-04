package actor_test

import (
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/actor"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTopologyPlanner(t *testing.T) {
	type key struct {
		name, locality string
	}

	type value struct {
		nodes int32
		join  string
	}

	tests := []struct {
		name     string
		topology *api.Topology
		nodes    int32
		expected map[key]value
	}{
		{
			name:     "single node cluster with locality not-set",
			topology: nil,
			nodes:    1,
			expected: map[key]value{
				key{"crdb", ""}: {1, "crdb-0.crdb.test-ns:26257"},
			},
		},
		{
			name:     "multinode cluster with locality not-set",
			topology: nil,
			nodes:    3,
			expected: map[key]value{
				key{"crdb", ""}: {3, "crdb-0.crdb.test-ns:26257,crdb-1.crdb.test-ns:26257,crdb-2.crdb.test-ns:26257"},
			},
		},
		{
			name: "with locality and equally splitting number of nodes (#1)",
			topology: &api.Topology{
				Zones: []api.AvailabilityZone{
					{
						StatefulSetSuffix: "-a",
						Locality:          "zone=zone-a",
					},
					{
						StatefulSetSuffix: "-b",
						Locality:          "zone=zone-b",
					},
					{
						StatefulSetSuffix: "-c",
						Locality:          "zone=zone-c",
					},
				},
			},
			nodes: 3,
			expected: map[key]value{
				key{"crdb-a", "zone=zone-a"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257"},
				key{"crdb-b", "zone=zone-b"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257"},
				key{"crdb-c", "zone=zone-c"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257"},
			},
		},
		{
			name: "with locality and unequal number of nodes (#2)",
			topology: &api.Topology{
				Zones: []api.AvailabilityZone{
					{
						StatefulSetSuffix: "-a",
						Locality:          "zone=zone-a",
					},
					{
						StatefulSetSuffix: "-b",
						Locality:          "zone=zone-b",
					},
				},
			},
			nodes: 2,
			expected: map[key]value{
				key{"crdb-a", "zone=zone-a"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257"},
				key{"crdb-b", "zone=zone-b"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257"},
			},
		},
		{
			name: "with locality and equally splitting number of nodes (#3)",
			topology: &api.Topology{
				Zones: []api.AvailabilityZone{
					{
						StatefulSetSuffix: "-a",
						Locality:          "zone=zone-a",
					},
					{
						StatefulSetSuffix: "-b",
						Locality:          "zone=zone-b",
					},
					{
						StatefulSetSuffix: "-c",
						Locality:          "zone=zone-c",
					},
					{
						StatefulSetSuffix: "-d",
						Locality:          "zone=zone-d",
					},
				},
			},
			nodes: 5,
			expected: map[key]value{
				key{"crdb-a", "zone=zone-a"}: {2, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257,crdb-d-0.crdb.test-ns:26257"},
				key{"crdb-b", "zone=zone-b"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257,crdb-d-0.crdb.test-ns:26257"},
				key{"crdb-c", "zone=zone-c"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257,crdb-d-0.crdb.test-ns:26257"},
				key{"crdb-d", "zone=zone-d"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257,crdb-d-0.crdb.test-ns:26257"},
			},
		},
		{
			name: "with locality and non-equal number of nodes (#1)",
			topology: &api.Topology{
				Zones: []api.AvailabilityZone{
					{
						StatefulSetSuffix: "-a",
						Locality:          "zone=zone-a",
					},
					{
						StatefulSetSuffix: "-b",
						Locality:          "zone=zone-b",
					},
					{
						StatefulSetSuffix: "-c",
						Locality:          "zone=zone-c",
					},
				},
			},
			nodes: 4,
			expected: map[key]value{
				key{"crdb-a", "zone=zone-a"}: {2, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257"},
				key{"crdb-b", "zone=zone-b"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257"},
				key{"crdb-c", "zone=zone-c"}: {1, "crdb-a-0.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257,crdb-c-0.crdb.test-ns:26257"},
			},
		},
		{
			name: "with locality and non-equal number of nodes (#2)",
			topology: &api.Topology{
				Zones: []api.AvailabilityZone{
					{
						StatefulSetSuffix: "-a",
						Locality:          "zone=zone-a",
					},
					{
						StatefulSetSuffix: "-b",
						Locality:          "zone=zone-b",
					},
				},
			},
			nodes: 5,
			expected: map[key]value{
				key{"crdb-a", "zone=zone-a"}: {3, "crdb-a-0.crdb.test-ns:26257,crdb-a-1.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257"},
				key{"crdb-b", "zone=zone-b"}: {2, "crdb-a-0.crdb.test-ns:26257,crdb-a-1.crdb.test-ns:26257,crdb-b-0.crdb.test-ns:26257"},
			},
		},
		{
			name: "with locality and non-equal number of nodes (#3)",
			topology: &api.Topology{
				Zones: []api.AvailabilityZone{
					{
						StatefulSetSuffix: "-a",
						Locality:          "zone=zone-a",
					},
					{
						StatefulSetSuffix: "-b",
						Locality:          "zone=zone-b",
					},
				},
			},
			nodes: 1,
			expected: map[key]value{
				key{"crdb-a", "zone=zone-a"}: {1, "crdb-a-0.crdb.test-ns:26257"},
				key{"crdb-b", "zone=zone-b"}: {0, "crdb-a-0.crdb.test-ns:26257"},
			},
		},
		{
			name: "with locality and zero nodes)",
			topology: &api.Topology{
				Zones: []api.AvailabilityZone{
					{
						StatefulSetSuffix: "-a",
						Locality:          "zone=zone-a",
					},
					{
						StatefulSetSuffix: "-b",
						Locality:          "zone=zone-b",
					},
				},
			},
			nodes: 0,
			expected: map[key]value{
				key{"crdb-a", "zone=zone-a"}: {0, ""},
				key{"crdb-b", "zone=zone-b"}: {0, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := testutil.NewBuilder("crdb").Namespaced("test-ns").WithTopology(tt.topology).WithNodeCount(tt.nodes).Cluster()

			planner := actor.TopologyPlanner{
				Cluster: cluster,
			}

			actual := make(map[key]value)

			require.NoError(t, planner.ForEachZone(func(name string, nodes int32, join string, loc string, nodeSelector map[string]string) error {
				actual[key{name, loc}] = value{nodes, join}

				return nil
			}))

			assert.Equal(t, tt.expected, actual)
		})
	}
}
