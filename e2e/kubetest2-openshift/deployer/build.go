/*
Copyright 2026 The Cockroach Authors

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

package deployer

import "errors"

// Build is used to deploy kubernetes to the cluster
// and is not supported by this deployer. It is required
// by the kubetest2 interface and cannot be used to deploy
// a built version of openshift.
func (d *deployer) Build() error {
	return errors.New("this deployer does not support building and deploying Kubernetes")
}
