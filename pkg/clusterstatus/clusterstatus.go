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

package clusterstatus

import (
	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetClusterStatusOnFirstReconcile will set Status of teh cluster with value Running on first rin
// Very important: this is used on OpenShift CI to pass the tests, if not set it will timeout
func SetClusterStatusOnFirstReconcile(status *api.CrdbClusterStatus) {
	InitOperatorActionsIfNeeded(status, metav1.Now())
	if status.ClusterStatus != "" {
		return
	}
	status.ClusterStatus = api.ActionStatus(api.Starting).String()
}

func SetClusterStatus(status *api.CrdbClusterStatus) {
	InitOperatorActionsIfNeeded(status, metav1.Now())
	for _, a := range status.OperatorActions {
		if a.Status == api.ActionStatus(api.Failed).String() {
			status.ClusterStatus = api.ActionStatus(api.Failed).String()
			return
		}

		if a.Status == api.ActionStatus(api.Unknown).String() {
			status.ClusterStatus = api.ActionStatus(api.Unknown).String()
			return
		}

		status.ClusterStatus = api.ActionStatus(api.Finished).String()
		return
	}
	status.ClusterStatus = api.ActionStatus(api.Finished).String()
}

func InitOperatorActionsIfNeeded(status *api.CrdbClusterStatus, now metav1.Time) {
	if status.OperatorActions == nil {
		status.OperatorActions = []api.ClusterAction{}
	}
}

func Finished(atype api.ActionType, actions []api.ClusterAction) bool {
	pos := pos(atype, actions)
	if pos == -1 || actions[pos].Status == api.ActionStatus(api.Unknown).String() {
		return false
	}

	return actions[pos].Status == api.ActionStatus(api.Finished).String()
}

func Failed(atype api.ActionType, actions []api.ClusterAction) bool {
	pos := pos(atype, actions)
	if pos == -1 || actions[pos].Status == api.ActionStatus(api.Unknown).String() {
		return false
	}

	return actions[pos].Status == api.ActionStatus(api.Failed).String()
}

func Unknown(atype api.ActionType, actions []api.ClusterAction) bool {
	pos := pos(atype, actions)
	if pos == -1 {
		return false
	}

	return actions[pos].Status == api.ActionStatus(api.Unknown).String()
}

func SetActionFailed(atype api.ActionType, message string, status *api.CrdbClusterStatus) {
	setActionStatus(atype, api.Failed, status, metav1.Now(), message)
	status.ClusterStatus = api.ActionStatus(api.Failed).String()
}

func SetActionFinished(atype api.ActionType, status *api.CrdbClusterStatus) {
	setActionStatus(atype, api.Finished, status, metav1.Now(), "")
}

func SetActionUnknown(atype api.ActionType, status *api.CrdbClusterStatus) {
	setActionStatus(atype, api.Unknown, status, metav1.Now(), "")
}

func setActionStatus(atype api.ActionType, status api.ActionStatus, clusterStatus *api.CrdbClusterStatus, now metav1.Time, message string) {
	action := findOrCreate(atype, clusterStatus, message)
	action.Message = message
	action.Status = status.String()
	action.LastTransitionTime = now
}

func findOrCreate(atype api.ActionType, status *api.CrdbClusterStatus, message string) *api.ClusterAction {
	pos := pos(atype, status.OperatorActions)
	if pos >= 0 {
		return &status.OperatorActions[pos]
	}

	status.OperatorActions = append(status.OperatorActions, api.ClusterAction{
		Type:               atype,
		Message:            message,
		Status:             api.ActionStatus(api.Unknown).String(),
		LastTransitionTime: metav1.Now(),
	})

	return &status.OperatorActions[len(status.OperatorActions)-1]
}

func pos(atype api.ActionType, actions []api.ClusterAction) int {
	for i := range actions {
		if actions[i].Type == atype {
			return i
		}
	}

	return -1
}
