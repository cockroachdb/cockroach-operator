package v1beta1

// ClusterConditionType is an enumeration of CrdbCluster conditions, listed
// below. Each condition should start with the "Cluster" prefix and be phrased
// as a state adjective, if possible.
type ClusterConditionType string

const (
	// ClusterInitialized is set to True once a newly created cluster has been
	// fully initialized and is ready for connections.
	ClusterInitialized ClusterConditionType = "Initialized"

	// ClusterRangesUnderReplicated is set to False once there are no
	// under-replicated ranges.
	ClusterRangesUnderReplicated ClusterConditionType = "RangesUnderReplicated"
)