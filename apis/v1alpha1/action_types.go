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

package v1alpha1

//ActionType type alias
type ActionType string

const (
	//VersionCheckerAction string
	VersionCheckerAction ActionType = "VersionCheckerAction"
	//ClusterRestartAction string
	ClusterRestartAction ActionType = "ClusterRestart"
	//DeployAction string
	DeployAction ActionType = "Deploy"
	//DecommissionAction string
	DecommissionAction ActionType = "Decommission"
	//InitializeAction string
	InitializeAction ActionType = "Initialize"
	//GenerateCert string
	GenerateCertAction ActionType = "GenerateCert"
	//RequestCertAction string
	ResizePVCAction ActionType = "ResizePVC"
	//PartitionedUpdateAction string
	PartitionedUpdateAction ActionType = "PartitionedUpdate"
	//UnknownAction string
	UnknownAction ActionType = "Unknown"
)
