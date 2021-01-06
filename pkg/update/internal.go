/*
Copyright 2021 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
