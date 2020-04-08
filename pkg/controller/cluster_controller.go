package controller

import (
	"context"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/actor"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// ClusterReconciler reconciles a CrdbCluster object
type ClusterReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Actions []actor.Actor
}

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log := r.Log.WithValues("CrdbCluster", req.NamespacedName)
	log.Info("reconciling CockroachDB cluster")

	fetcher := resource.NewKubeFetcher(ctx, req.Namespace, r.Client)

	cr := resource.ClusterPlaceholder(req.Name)
	if err := fetcher.Fetch(cr); err != nil {
		log.Error(err, "failed to retrieve CrdbCluster resource")
		return requeueIfError(client.IgnoreNotFound(err))
	}

	cluster := resource.NewCluster(cr)

	ctx = actor.ContextWithCancelFn(ctx, cancel)

	for _, a := range r.Actions {
		if a.Handles(cluster.Status().Conditions) {
			if err := a.Act(ctx, &cluster); err != nil {
				if notReadyErr, ok := err.(actor.NotReadyErr); ok {
					log.Info("requeuing", "reason", notReadyErr.Error())
					return requeueImmediately()
				}

				log.Error(err, "action failed")
				return requeueIfError(err)
			}
		}

		if cancelled(ctx) {
			log.Info("request was interrupted")
			return noRequeue()
		}
	}

	fresh, err := cluster.IsFresh(fetcher)
	if err != nil {
		return requeueIfError(err)
	}

	if !fresh {
		log.Info("cluster resources is not up to date")
		return requeueImmediately()
	}

	if err := r.Client.Status().Update(ctx, cluster.Unwrap()); err != nil {
		log.Error(err, "failed to update cluster status")
		return requeueIfError(err)
	}

	log.Info("reconciliation completed")
	return noRequeue()
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.CrdbCluster{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}

func InitClusterReconciler() func(ctrl.Manager) error {
	return InitClusterReconcilerWithLogger(ctrl.Log.WithName("controller").WithName("CrdbCluster"))
}

func InitClusterReconcilerWithLogger(l logr.Logger) func(ctrl.Manager) error {
	return func(mgr ctrl.Manager) error {
		return (&ClusterReconciler{
			Client:  mgr.GetClient(),
			Log:     l,
			Scheme:  mgr.GetScheme(),
			Actions: actor.NewOperatorActions(mgr.GetScheme(), mgr.GetClient(), mgr.GetConfig()),
		}).SetupWithManager(mgr)
	}
}

func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
