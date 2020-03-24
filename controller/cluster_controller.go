package controller

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crdbv1alpha1 "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
)

// CrdbClusterReconciler reconciles a CrdbCluster object
type CrdbClusterReconciler struct {
	client.Client
	Log logr.Logger
	Scheme *runtime.Scheme
}

func (r *CrdbClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("CrdbCluster", req.NamespacedName)

	log.Info("reconciling CockroachDB cluster")

	cluster := &crdbv1alpha1.CrdbCluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		log.Error(err, "failed to retrieve CrdbCluster resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

func (r *CrdbClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdbv1alpha1.CrdbCluster{}).
		Complete(r)
}
