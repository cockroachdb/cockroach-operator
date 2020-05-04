package condition

import (
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitConditionsIfNeeded(status *api.CrdbClusterStatus, now metav1.Time) {
	if status.Conditions == nil {
		status.Conditions = []api.ClusterCondition{}
	}

	if len(status.Conditions) == 0 {
		SetFalse(api.InitializedCondition, status, now)
	}
}

func False(ctype api.ClusterConditionType, conds []api.ClusterCondition) bool {
	pos := pos(ctype, conds)
	if pos == -1 || conds[pos].Status == metav1.ConditionUnknown {
		return false
	}

	return conds[pos].Status == metav1.ConditionFalse
}

func Unknown(ctype api.ClusterConditionType, conds []api.ClusterCondition) bool {
	pos := pos(ctype, conds)
	if pos == -1 {
		return false
	}

	return conds[pos].Status == metav1.ConditionUnknown
}

func SetFalse(ctype api.ClusterConditionType, status *api.CrdbClusterStatus, now metav1.Time) {
	setStatus(ctype, metav1.ConditionFalse, status, now)
}

func SetTrue(ctype api.ClusterConditionType, status *api.CrdbClusterStatus, now metav1.Time) {
	setStatus(ctype, metav1.ConditionTrue, status, now)
}

func setStatus(ctype api.ClusterConditionType, status metav1.ConditionStatus, clusterStatus *api.CrdbClusterStatus, now metav1.Time) {
	cond := findOrCreate(ctype, clusterStatus)

	if cond.Status == status {
		return
	}

	cond.Status = status
	cond.LastTransitionTime = now
}

func findOrCreate(ctype api.ClusterConditionType, status *api.CrdbClusterStatus) *api.ClusterCondition {
	pos := pos(ctype, status.Conditions)
	if pos >= 0 {
		return &status.Conditions[pos]
	}

	status.Conditions = append(status.Conditions, api.ClusterCondition{
		Type:               ctype,
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.Now(),
	})

	return &status.Conditions[len(status.Conditions)-1]
}

func pos(ctype api.ClusterConditionType, conds []api.ClusterCondition) int {
	for i := range conds {
		if conds[i].Type == ctype {
			return i
		}
	}

	return -1
}
