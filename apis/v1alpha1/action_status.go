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

//ActionStatus struct used to save the status of each action that the controller runs
type ActionStatus int

const (
	//Failed status
	Failed = iota
	//Starting status
	Starting
	//Finished status
	Finished
	//Unknown status
	Unknown
)

var statuses []string = []string{
	"Failed",
	"Starting",
	"Finished",
	"Unknown",
}

func (a ActionStatus) String() string {
	if a < Failed || a > Unknown {
		return "Unknown"
	}
	return statuses[a]
}
