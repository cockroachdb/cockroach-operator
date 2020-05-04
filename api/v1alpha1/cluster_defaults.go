package v1alpha1

var (
	DefaultGRPCPort int32 = 26257
	DefaultHTTPPort int32 = 8080
)

func SetClusterSpecDefaults(cs *CrdbClusterSpec) {
	if cs.GRPCPort == nil {
		cs.GRPCPort = &DefaultGRPCPort
	}

	if cs.HTTPPort == nil {
		cs.HTTPPort = &DefaultHTTPPort
	}

	if cs.Topology == nil {
		cs.Topology = &Topology{}
	}

	if len(cs.Topology.Zones) == 0 {
		cs.Topology.Zones = []AvailabilityZone{{}}
	}
}
