package actor

import (
	"context"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newScaleStatus(scheme *runtime.Scheme, cl client.Client, config *rest.Config, clientset kubernetes.Interface) Actor {
	return &scaleStatus{
		action: newAction(scheme, cl, config, clientset),
	}
}

// This actor gets stable replicas and labelselector from statefulset subresource
type scaleStatus struct {
	action
}

// GetActionType returns api.scalstatus action used to set the cluster status errors
func (sss *scaleStatus) GetActionType() api.ActionType {
	return api.ScaleStatusAction
}

// Act in this implementation get replicas and labelselector of a CR sts.
func (sss *scaleStatus) Act(ctx context.Context, cluster *resource.Cluster, log logr.Logger) error {

	// Get the sts and compare the sts size to the size in the CR
	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      cluster.StatefulSetName(),
	}
	statefulSet := &appsv1.StatefulSet{}
	if err := sss.client.Get(ctx, key, statefulSet); err != nil {
		return errors.Wrap(err, "failed to fetch statefulset")
	}

	status := &statefulSet.Status
	if status.CurrentReplicas == 0 || status.CurrentReplicas < status.Replicas {
		log.Info("Statefulset is not healthy")
		return NotReadyErr{Err: errors.New("Statefulset is not healthy")}
	}

	cluster.SetClusterNodes(status.CurrentReplicas)
	cluster.SetClusterSelector(metav1.FormatLabelSelector(statefulSet.Spec.Selector))

	// log.Info("started scale status now")

	// if err := sss.client.Status().Update(ctx, cluster.Unwrap()); err != nil {
	// 	log.Error(err, "failed saving cluster status on replicas and selector")
	// 	return err
	// }

	return nil
}
