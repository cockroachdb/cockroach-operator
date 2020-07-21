/*
Copyright 2020 The Cockroach Authors

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

package controller

import (
	"context"
	"time"

	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterReconciler reconciles a CrdbCluster object
type ClusterReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Actions []actor.Actor
}

// Note: you need a blank line after this list in order for the controller to pick this up.

// +kubebuilder:rbac:groups=crdb.cockroachlabs.com,resources=crdbclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crdb.cockroachlabs.com,resources=crdbclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch;create;patch;delete
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/approval,verbs=update
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get
// +kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets/status,verbs=get

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
					log.Info("requeueing", "reason", notReadyErr.Error())
					return requeueAfter(5*time.Second, nil)
				}

				if cantRecoverErr, ok := err.(actor.PermanentErr); ok {
					log.Error(cantRecoverErr, "can't proceed with reconcile")
					return requeueAfter(5*time.Minute, err)
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
		Owns(&policy.PodDisruptionBudget{}).
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
