/*
Copyright 2025 The Cockroach Authors

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

package v1alpha1

// ClusterConditionType type alias
type ClusterConditionType string

const (
	//CrdbCertificateGenerated condition used to run the version checker and sync other actions
	CertificateGenerated ClusterConditionType = "CertificateGenerated"
	//CrdbVersionChecked condition used to run the version checker and sync other actions
	CrdbVersionChecked ClusterConditionType = "CrdbVersionChecked"
	//DecommissionCondition string
	DecommissionCondition ClusterConditionType = "Decommission"
	//CrdbInitializedCondition string
	CrdbInitializedCondition ClusterConditionType = "Initialized"
	// CrdbUIIngressExposedCondition string
	CrdbUIIngressExposedCondition ClusterConditionType = "UIIngressExposed"
	// CrdbSQLIngressExposedCondition string
	CrdbSQLIngressExposedCondition ClusterConditionType = "SQLIngressExposed"
	//ClusterRestartCondition string
	ClusterRestartCondition ClusterConditionType = "RestartedCluster"
)
