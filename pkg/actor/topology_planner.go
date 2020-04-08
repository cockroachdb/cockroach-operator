package actor

import (
	"fmt"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"strings"
)

type TopologyPlanner struct {
	Cluster *resource.Cluster
}

type AZUpdater func(name string, nodes int32, joinStr string, locality string, nodeSelector map[string]string) error

func (p TopologyPlanner) ForEachZone(updater AZUpdater) error {
	topology := p.Cluster.Spec().Topology.DeepCopy()

	stsNames := make([]string, len(topology.Zones))
	for i, zone := range topology.Zones {
		stsNames[i] = p.Cluster.StatefulSetName() + zone.StatefulSetSuffix
	}

	buckets, joinStr := p.plan(int(p.Cluster.Spec().Nodes), stsNames, topology.Zones)

	for i, zone := range topology.Zones {
		err := updater(stsNames[i], buckets[i], joinStr, zone.Locality, topology.Zones[i].Labels)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p TopologyPlanner) plan(nodes int, stsNames []string, zones []api.AvailabilityZone) ([]int32, string) {
	numZones := len(zones)
	buckets := make([]int32, numZones)

	var seeds []string

	for i := 0; i < numZones; i++ {
		buckets[i] = int32(nodes / numZones)

		node := func(id int) string {
			node := fmt.Sprintf("%s-%d", stsNames[i], id)
			return fmt.Sprintf("%s.%s.%s", node, stsNames[i], p.Cluster.Namespace())
		}

		// All first nodes go into the seeds list
		if buckets[i] > 0 {
			seeds = append(seeds, fmt.Sprintf("%s:%d", node(0), *p.Cluster.Spec().GRPCPort))
		}

		reminder := nodes % numZones

		if numZones == 1 {
			for j := 1; j < nodes && len(seeds) < 3; j++ {
				seeds = append(seeds, fmt.Sprintf("%s:%d", node(j), *p.Cluster.Spec().GRPCPort))
			}
		}

		if i < reminder {
			// this condition needs to work for the cases up to 2 zones and 3 nodes
			if len(seeds) < 3 && numZones < 3 && len(seeds) < nodes && buckets[i] <= 2 {
				idx := 0
				if len(seeds) > 0 {
					idx = 1
				}
				seeds = append(seeds, fmt.Sprintf("%s:%d", node(idx), *p.Cluster.Spec().GRPCPort))
			}
			buckets[i]++
		}
	}
	return buckets, strings.Join(seeds, ",")
}
