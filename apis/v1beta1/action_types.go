package v1beta1

type ActionType string

const (
	// Upgrade action type represents the upgrade of the cockroach version
	// running on the cluster.
	//
	// For host clusters, when Upgrade is set to Finished, we can be sure that
	// all tenants are running an up-to-date CRDB image.
	Upgrade ActionType = "Upgrade"
	// AwaitFinalization action type represents a post upgrade but
	// pre-finalization state for the cluster.
	AwaitFinalization ActionType = "AwaitFinalization"

	// ValidateVersion action type represents the validation of the cockroach version
	// running on the cluster.
	ValidateVersion ActionType = "ValidateVersion"
)