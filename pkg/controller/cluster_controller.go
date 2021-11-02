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
	"fmt"
	"time"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/actor"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/go-logr/logr"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ClusterReconciler reconciles a CrdbCluster object
type ClusterReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Director actor.Director
}

// Note: you need a blank line after this list in order for the controller to pick this up.

// +kubebuilder:rbac:groups=crdb.cockroachlabs.com,resources=crdbclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crdb.cockroachlabs.com,resources=crdbclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=crdb.cockroachlabs.com,resources=crdbclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests,verbs=get;list;watch;create;patch;delete
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests/approval,verbs=update
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get
// +kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups=core,resources=pods/log,verbs=get
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;create;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;create;watch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;create;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/scale,verbs=get;watch;update
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets/status,verbs=get
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups=batch,resources=jobs/finalizers,verbs=get;list;watch
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;update;patch;
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;update;patch;
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=use

// Reconcile is the reconciliation loop entry point for cluster CRDs.  It fetches the current cluster resources
// and uses its state to interact with the world via a set of actions implemented by `Actor`s
// (i.e. init cluster, create a statefulset).
// Each action can result in:
//   - a short requeue (5 seconds)
//   - a long requeue (5 minutes)
//   - cancel the loop and wait for another event
//   - if no other errors occurred continue to the next action
func (r *ClusterReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	// TODO should we make this configurable?
	// Ensure the loop does not take longer than 4 hours
	ctx, cancel := context.WithTimeout(ctx, 10*time.Hour)
	defer cancel()

	log := r.Log.WithValues("CrdbCluster", req.NamespacedName, "ReconcileId", shortuuid.New())
	log.V(int(zapcore.InfoLevel)).Info("reconciling CockroachDB cluster")

	fetcher := resource.NewKubeFetcher(ctx, req.Namespace, r.Client)

	cr := resource.ClusterPlaceholder(req.Name)
	if err := fetcher.Fetch(cr); err != nil {
		log.Error(err, "failed to retrieve CrdbCluster resource")
		return requeueIfError(client.IgnoreNotFound(err))
	}

	cluster := resource.NewCluster(cr)
	// on first run we need to save the status and exit to pass Openshift CI
	// we added a state called Starting for field ClusterStatus to accomplish this
	if cluster.Status().ClusterStatus == "" {
		cluster.SetClusterStatusOnFirstReconcile()
		if err := r.Client.Status().Update(ctx, cluster.Unwrap()); err != nil {
			log.Error(err, "failed to update cluster status on action")
			return requeueIfError(err)
		}
		return requeueImmediately()
	}

	if cluster.Status().DirectorState == "" {
		log.V(int(zapcore.InfoLevel)).Info("initializing director state")
		cluster.UpdateDirectorState(actor.DirectorStateAvailable, "")
		if err := r.Client.Status().Update(ctx, cluster.Unwrap()); err != nil {
			log.Error(err, "failed to initialize director state")
			return requeueIfError(err)
		}
		return requeueImmediately()
	}

	//force version validation on mismatch between status and spec
	if cluster.True(api.CrdbVersionChecked) {
		if cluster.GetCockroachDBImageName() != cluster.Status().CrdbContainerImage {
			cluster.SetFalse(api.CrdbVersionChecked)
			if err := r.Client.Status().Update(ctx, cluster.Unwrap()); err != nil {
				log.Error(err, "failed to update cluster status on action")
				return requeueIfError(err)
			}
			return requeueImmediately()
		}
	}

	actorToExecute, err := r.Director.GetActorToExecute(ctx, &cluster, log)
	if err != nil {
		return requeueAfter(30*time.Second, nil)
	} else if actorToExecute == nil {
		log.Info("No actor to run; not requeueing")
		return noRequeue()
	}

	// Save context cancellation function for actors to call if needed
	ctx = actor.ContextWithCancelFn(ctx, cancel)

	log.Info(fmt.Sprintf("Running action with name: %s", actorToExecute.GetActionType()))
	if err := actorToExecute.Act(ctx, &cluster, log); err != nil {
		// Save the error on the Status for each action
		log.Info("Error on action", "Action", actorToExecute.GetActionType(), "err", err.Error())
		cluster.SetActionFailed(actorToExecute.GetActionType(), err.Error())
		defer func(ctx context.Context, cluster *resource.Cluster) {
			if err := r.Client.Status().Update(ctx, cluster.Unwrap()); err != nil {
				log.Error(err, "failed to update cluster status")
			}
		}(ctx, &cluster)
		// Short pause
		if notReadyErr, ok := err.(actor.NotReadyErr); ok {
			log.V(int(zapcore.DebugLevel)).Info("requeueing", "reason", notReadyErr.Error(), "Action", actorToExecute.GetActionType())
			return requeueAfter(5*time.Second, nil)
		}

		// Long pause
		if cantRecoverErr, ok := err.(actor.PermanentErr); ok {
			log.Error(cantRecoverErr, "can't proceed with reconcile", "Action", actorToExecute.GetActionType())
			return noRequeue()
		}

		// No requeue until the user makes changes
		if validationError, ok := err.(actor.ValidationError); ok {
			log.Error(validationError, "can't proceed with reconcile")
			return noRequeue()
		}

		log.Error(err, "action failed")
		return requeueIfError(err)
	}
	// reset errors on each run  if there was an error,
	// this is to cover the not ready case
	if cluster.Failed(actorToExecute.GetActionType()) {
		cluster.SetActionFinished(actorToExecute.GetActionType())
	}

	// Stop processing and wait for Kubernetes scheduler to call us again as the actor
	// modified actorToExecute resource owned by the controller
	if cancelled(ctx) {
		log.V(int(zapcore.InfoLevel)).Info("request was interrupted")
		return noRequeue()
	}

	refreshedCr := resource.ClusterPlaceholder(cluster.Name())
	if err := fetcher.Fetch(refreshedCr); err != nil {
		log.Error(err, "failed to retrieve refreshed CrdbCluster resource")
		return requeueIfError(client.IgnoreNotFound(err))
	}

	// Check if the resource has been updated while the controller worked on it
	specChanged, err := cluster.SpecChanged(refreshedCr)
	if err != nil {
		return requeueIfError(err)
	}
	// If the resource was updated, it is needed to start all over again
	// to ensure that the latest state was reconciled
	if specChanged {
		log.V(int(zapcore.DebugLevel)).Info("cluster resources is not up to date")
		return requeueImmediately()
	}

	refreshedCluster := resource.NewCluster(refreshedCr)
	refreshedCluster.SetClusterStatus()
	if err := r.Client.Status().Update(ctx, refreshedCluster.Unwrap()); err != nil {
		log.Error(err, "failed to update cluster status")
		return requeueIfError(err)
	}

	log.V(int(zapcore.InfoLevel)).Info("reconciliation completed")
	return noRequeue()
}

// SetupWithManager registers the controller with the controller.Manager from controller-runtime
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.CrdbCluster{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&policy.PodDisruptionBudget{}).
		Owns(&kbatch.Job{}).
		Complete(r)
}

// InitClusterReconciler returns a registrator for new controller instance with the default logger
func InitClusterReconciler() func(ctrl.Manager) error {
	return InitClusterReconcilerWithLogger(ctrl.Log.WithName("controller").WithName("CrdbCluster"))
}

// InitClusterReconcilerWithLogger returns a registrator for new controller instance with provided logger
func InitClusterReconcilerWithLogger(l logr.Logger) func(ctrl.Manager) error {
	return func(mgr ctrl.Manager) error {
		clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
		if err != nil {
			return err
		}
		return (&ClusterReconciler{
			Client:   mgr.GetClient(),
			Log:      l,
			Scheme:   mgr.GetScheme(),
			Director: actor.NewDirector(mgr.GetScheme(), mgr.GetClient(), mgr.GetConfig(), clientset),
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
