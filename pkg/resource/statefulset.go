package resource

import (
	"context"
	crdbv1alpha1 "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/label"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSet struct {
	Cluster *crdbv1alpha1.CrdbCluster
}

func (ss *StatefulSet) Reconcile(ctx context.Context, cl client.Client) (*appsv1.StatefulSet, error) {
	existing := &appsv1.StatefulSet{}

	key := ss.makeNamespacedName()
	if err := cl.Get(ctx, key, existing); client.IgnoreNotFound(err) != nil {
		return nil, errors.Wrapf(err, "failed to fetch statefulset: %s", key)
	}

	desired := ss.makeDesired(key)

	if equalStatefulSet(existing, desired) {
		return nil, nil
	}

	return desired, nil
}

func (ss *StatefulSet) makeDesired(nn types.NamespacedName) *appsv1.StatefulSet {
	meta := metav1.ObjectMeta{
		Name:      nn.Name,
		Namespace: nn.Namespace,
		Labels:    label.MakeCommonLabels(ss.Cluster),
	}

	desired := &appsv1.StatefulSet{
		ObjectMeta: meta,
		Spec:       *ss.Cluster.Spec.NodesSpec.DeepCopy(),
	}

	copyLabel(label.NameLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Selector.MatchLabels)
	copyLabel(label.InstanceLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Selector.MatchLabels)
	copyLabel(label.ComponentLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Selector.MatchLabels)

	copyLabel(label.NameLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Template.ObjectMeta.Labels)
	copyLabel(label.InstanceLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Template.ObjectMeta.Labels)
	copyLabel(label.ComponentLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Template.ObjectMeta.Labels)

	copyLabel(label.NameLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels)
	copyLabel(label.InstanceLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels)
	copyLabel(label.ComponentLabelKey, ss.Cluster.ObjectMeta.Labels, desired.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels)

	return desired
}

func (ss *StatefulSet) makeNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      "cockroachdb",
		Namespace: ss.Cluster.GetNamespace(),
	}
}

func equalStatefulSet(l *appsv1.StatefulSet, r *appsv1.StatefulSet) bool {
	return cmp.Equal(l.ObjectMeta.Labels, r.ObjectMeta.Labels) &&
		cmp.Equal(l.ObjectMeta.Annotations, r.ObjectMeta.Annotations) &&
		cmp.Equal(l.Spec, r.Spec)
}

func copyLabel(key string, from map[string]string, to map[string]string) {
	to[key] = from[key]
}
