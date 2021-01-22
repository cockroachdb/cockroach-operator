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
// +kubebuilder:rbac:groups=core,resources=services/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get
// +kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets/status,verbs=get
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets/finalizers,verbs=get;list;watch

// Reconcile is the reconciliation loop entry point for cluster CRDs.  It fetches the current cluster resources
// and uses its state to interact with the world via a set of actions implemented by `Actor`s
// (i.e. init cluster, create a statefulset).
// Each action can result in:
//   - a short requeue (5 seconds)
//   - a long requeue (5 minutes)
//   - cancel the loop and wait for another event
//   - if no other errors occurred continue to the next action
func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	// TODO should we make this configurable?
	// Ensure the loop does not take longer than 4 hours
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Hour)
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
	


	// Save context cancellation function for actors to call if needed
	ctx = actor.ContextWithCancelFn(ctx, cancel)

	// Apply all actions to the cluster. Some actions can stop the loop if it is needed
	// to refresh the state of the world
	for _, a := range r.Actions {
		// Ensure the action is applicable to the current resource state
		if a.Handles(cluster.Status().Conditions) {
			if err := a.Act(ctx, &cluster); err != nil {
				// Short pause
				if notReadyErr, ok := err.(actor.NotReadyErr); ok {
					log.Info("requeueing", "reason", notReadyErr.Error())
					return requeueAfter(5*time.Second, nil)
				}

				// Long pause
				if cantRecoverErr, ok := err.(actor.PermanentErr); ok {
					log.Error(cantRecoverErr, "can't proceed with reconcile")
					return requeueAfter(5*time.Minute, err)
				}

				log.Error(err, "action failed")
				return requeueIfError(err)
			}
		}

		// Stop processing and wait for Kubernetes scheduler to call us again as the actor
		// modified a resource owned by the controller
		if cancelled(ctx) {
			log.Info("request was interrupted")
			return noRequeue()
		}
	}

	// Check if the resource has been updated while the controller worked on it
	fresh, err := cluster.IsFresh(fetcher)
	if err != nil {
		return requeueIfError(err)
	}

	// If the resource was updated, it is needed to start all over again
	// to ensure that the latest state was reconciled
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

// SetupWithManager registers the controller with the controller.Manager from controller-runtime
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.CrdbCluster{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&policy.PodDisruptionBudget{}).
		Complete(r)
}

// InitClusterReconciler returns a registrator for new controller instance with the default logger
func InitClusterReconciler() func(ctrl.Manager) error {
	return InitClusterReconcilerWithLogger(ctrl.Log.WithName("controller").WithName("CrdbCluster"))
}

// InitClusterReconcilerWithLogger returns a registrator for new controller instance with provided logger
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
