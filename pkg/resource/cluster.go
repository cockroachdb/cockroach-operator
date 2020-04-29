package resource

import (
	"fmt"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/condition"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func NewCluster(original *api.CrdbCluster) Cluster {
	cr := original.DeepCopy()

	api.SetClusterSpecDefaults(&cr.Spec)

	timeNow := metav1.Now()
	condition.InitConditionsIfNeeded(&cr.Status, timeNow)

	return Cluster{
		cr:       cr,
		initTime: timeNow,
	}
}

func ClusterPlaceholder(name string) *api.CrdbCluster {
	return &api.CrdbCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

type Cluster struct {
	Fetcher

	cr       *api.CrdbCluster
	scheme   *runtime.Scheme
	initTime metav1.Time
}

func (cluster Cluster) Unwrap() *api.CrdbCluster {
	return cluster.cr.DeepCopy()
}

func (cluster Cluster) SetTrue(ctype api.ClusterConditionType) {
	condition.SetTrue(ctype, &cluster.cr.Status, cluster.InitTime())
}

func (cluster Cluster) Spec() *api.CrdbClusterSpec {
	return cluster.cr.Spec.DeepCopy()
}

func (cluster Cluster) Status() *api.CrdbClusterStatus {
	return cluster.cr.Status.DeepCopy()
}

func (cluster Cluster) Name() string {
	return cluster.cr.Name
}

func (cluster Cluster) Namespace() string {
	return cluster.cr.Namespace
}

func (cluster Cluster) ObjectKey() types.NamespacedName {
	return types.NamespacedName{Namespace: cluster.Namespace(), Name: cluster.Name()}
}

func (cluster Cluster) InitTime() metav1.Time {
	return cluster.initTime
}

func (cluster Cluster) DiscoveryServiceName() string {
	return cluster.Name()
}

func (cluster Cluster) PublicServiceName() string {
	return fmt.Sprintf("%s-public", cluster.Name())
}

func (cluster Cluster) StatefulSetName() string {
	return cluster.Name()
}

func (cluster Cluster) IsFresh(fetcher Fetcher) (bool, error) {
	actual := ClusterPlaceholder(cluster.Name())
	if err := fetcher.Fetch(actual); err != nil {
		return false, errors.Wrapf(err, "failed to fetch cluster resource")
	}

	return cluster.cr.ResourceVersion == actual.ResourceVersion, nil
}
