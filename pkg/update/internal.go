package update

// Cockroach roles that we can grant/revoke for users.
const (
	// Fixed fields in the client certificate. Any other values will be rejected by Vault.
	// Remember to update <repo root>/conf/policies/intrusion.hcl when updating this list of users.
	RootSQLUser = "root"
	NodeUser    = "node"
	AdminRole   = "admin"
)

// internalUsers is a set of SQL users created as part of the managed service, not to be used
// by customers. This struct is used to hide specific users in the console.
var internalUsers = map[string]struct{}{
	RootSQLUser: {},
	NodeUser:    {},
}

// internalDBs is a set of SQL databases created as part of CRDB, not to be used by customers.
var internalDBs = map[string]struct{}{
	"system":   {},
	"postgres": {},
}

func IsInternalUser(user string) bool {
	_, ok := internalUsers[user]
	return ok
}

func IsInternalDB(db string) bool {
	_, ok := internalDBs[db]
	return ok
}
