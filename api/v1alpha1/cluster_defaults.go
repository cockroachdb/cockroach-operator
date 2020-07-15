package v1alpha1

var (
	DefaultGRPCPort       int32 = 26257
	DefaultHTTPPort       int32 = 8080
	DefaultMaxUnavailable int32 = 1
)

func SetClusterSpecDefaults(cs *CrdbClusterSpec) {
	if cs.GRPCPort == nil {
		cs.GRPCPort = &DefaultGRPCPort
	}

	if cs.HTTPPort == nil {
		cs.HTTPPort = &DefaultHTTPPort
	}

	if cs.MaxUnavailable == nil {
		cs.GRPCPort = &DefaultMaxUnavailable
	}

	if cs.Cache == "" {
		cs.Cache = "25%"
	}

	if cs.MaxSQLMemory == "" {
		cs.MaxSQLMemory = "25%"
	}
}
