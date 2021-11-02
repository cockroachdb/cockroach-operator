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

package actor

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"time"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DirectorStateAvailable = "available"
	DirectorStateBusy      = "busy"

	maxTimeBusy = 12 * time.Hour
)

// DirectorLockError indicates that the director is busy when an actor tries to act
type DirectorLockError struct {
	Err error
}

func (e DirectorLockError) Error() string {
	return e.Err.Error()
}

type Director interface {
	GetActor(api.ActionType) Actor
	GetActorToExecute(context.Context, *resource.Cluster, logr.Logger) (Actor, error)
	ActAtomically(context.Context, *resource.Cluster, Actor, logr.Logger) error
}

type clusterDirector struct {
	actors     map[api.ActionType]Actor
	client     client.Client
	clientset  kubernetes.Interface
	scheme     *runtime.Scheme
	kubeDistro kube.KubernetesDistribution
	config     *rest.Config
}

func NewDirector(scheme *runtime.Scheme, cl client.Client, config *rest.Config, clientset kubernetes.Interface) Director {
	kd := kube.NewKubernetesDistribution()
	actors := map[api.ActionType]Actor{
		api.ClusterRestartAction:    newClusterRestart(cl, config, clientset),
		api.SetupRBACAction:         newSetupRBACAction(scheme, cl),
		api.DecommissionAction:      newDecommission(cl, config, clientset),
		api.VersionCheckerAction:    newVersionChecker(scheme, cl, clientset),
		api.GenerateCertAction:      newGenerateCert(cl),
		api.PartitionedUpdateAction: newPartitionedUpdate(cl, config, clientset),
		api.ResizePVCAction:         newResizePVC(scheme, cl, clientset),
		api.DeployAction:            newDeploy(scheme, cl, kd, clientset),
		api.InitializeAction:        newInitialize(scheme, cl, config, clientset),
	}
	return &clusterDirector{
		actors:     actors,
		client:     cl,
		clientset:  clientset,
		scheme:     scheme,
		kubeDistro: kd,
		config:     config,
	}
}

func (cd *clusterDirector) GetActor(aType api.ActionType) Actor {
	return cd.actors[aType]
}

func (cd *clusterDirector) GetActorToExecute(ctx context.Context, cluster *resource.Cluster, log logr.Logger) (Actor, error) {
	if cd.needsRestart(cluster) {
		return cd.actors[api.ClusterRestartAction], nil
	}

	needsRBACSetup, err := cd.needsRBACSetup(cluster)
	if err != nil {
		return nil, err
	} else if needsRBACSetup {
		return cd.actors[api.SetupRBACAction], nil
	}

	stsKey := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      cluster.StatefulSetName(),
	}
	ss := &appsv1.StatefulSet{}
	err = kube.IgnoreNotFound(cd.client.Get(ctx, stsKey, ss))
	if err != nil {
		return nil, err
	}

	if cd.needsDecommission(cluster, ss) {
		return cd.actors[api.DecommissionAction], nil
	}

	if cd.needsVersionCheck(cluster) {
		return cd.actors[api.VersionCheckerAction], nil
	}

	if cd.needsCertificateGeneration(cluster) {
		return cd.actors[api.GenerateCertAction], nil
	}

	if cd.needsPartitionedUpdate(cluster, ss) {
		return cd.actors[api.PartitionedUpdateAction], nil
	}

	if cd.needsPVCResize(cluster, ss) {
		return cd.actors[api.ResizePVCAction], nil
	}

	needsDeploy, err := cd.needsDeploy(ctx, cluster, log)
	if err != nil {
		return nil, err
	} else if needsDeploy {
		return cd.actors[api.DeployAction], nil
	}

	if cd.needsInitialization(cluster) {
		return cd.actors[api.InitializeAction], nil
	}

	return nil, nil
}

func (cd *clusterDirector) needsRestart(cluster *resource.Cluster) bool {
	conditions := cluster.Status().Conditions
	featureClusterRestartEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.ClusterRestart)
	featureVersionValidatorEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
	conditionVersionCheckedTrue := condition.True(api.CrdbVersionChecked, conditions)
	conditionInitializedTrue := condition.True(api.CrdbInitializedCondition, conditions)
	conditionInitializedFalse := condition.False(api.CrdbInitializedCondition, conditions)

	// In order to restart,
	// - the cluster restart feature must be enabled
	// - the version validator feature must be enabled and the version must be checked
	// - the cluster initialized condition must be true or false (not unknown)
	// - the cluster must have a restart annotation set

	if !featureClusterRestartEnabled {
		return false
	}
	if !featureVersionValidatorEnabled {
		return false
	}
	if !conditionVersionCheckedTrue {
		return false
	}
	if !conditionInitializedTrue && !conditionInitializedFalse {
		return false
	}

	return cluster.GetAnnotationRestartType() != ""
}

func (cd *clusterDirector) needsRBACSetup(cluster *resource.Cluster) (bool, error) {
	serviceAccounts := cd.clientset.CoreV1().ServiceAccounts(cluster.Namespace())
	if _, err := serviceAccounts.Get(context.Background(), cluster.ServiceAccountName(), metav1.GetOptions{}); kube.IsNotFound(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}
	return false, nil
}

func (cd *clusterDirector) needsDecommission(cluster *resource.Cluster, ss *appsv1.StatefulSet) bool {
	conditions := cluster.Status().Conditions
	featureDecommissionEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.Decommission)
	conditionInitializedTrue := condition.True(api.CrdbInitializedCondition, conditions)

	// In order to decommission,
	// - the decommission feature must be enabled
	// - the cluster must be initialized
	// - the current number of nodes must match the previously specified number of nodes, and that number must exceed the
	//   currently specified number of nodes

	if !featureDecommissionEnabled {
		return false
	}
	if !conditionInitializedTrue {
		return false
	}

	status := &ss.Status
	return status.CurrentReplicas == status.Replicas && status.CurrentReplicas > cluster.Spec().Nodes
}

func (cd *clusterDirector) needsVersionCheck(cluster *resource.Cluster) bool {
	conditions := cluster.Status().Conditions
	featureVersionValidatorEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
	conditionInitializedTrue := condition.True(api.CrdbInitializedCondition, conditions)
	conditionInitializedFalse := condition.False(api.CrdbInitializedCondition, conditions)
	conditionVersionCheckedFalse := condition.False(api.CrdbVersionChecked, conditions)

	// In order to check the version of the cluster,
	// - the version validator feature must be enabled
	// - the version should not already be checked
	// - the cluster initialized condition must be true or false (not unknown)

	if !featureVersionValidatorEnabled {
		return false
	}
	if !conditionVersionCheckedFalse {
		return false
	}
	if !conditionInitializedTrue && !conditionInitializedFalse {
		return false
	}
	return true
}

func (cd *clusterDirector) needsCertificateGeneration(cluster *resource.Cluster) bool {
	conditions := cluster.Status().Conditions
	conditionCertificateGeneratedTrue := condition.True(api.CertificateGenerated, conditions)

	// In order to generate a certificate,
	// - the certificate should not already be generated
	// - TLS should be enabled and a certificate should not already be provided

	if conditionCertificateGeneratedTrue {
		return false
	}

	return cluster.Spec().TLSEnabled && cluster.Spec().NodeTLSSecret == ""
}

func (cd *clusterDirector) needsPartitionedUpdate(cluster *resource.Cluster, ss *appsv1.StatefulSet) bool {
	conditions := cluster.Status().Conditions
	featureVersionValidatorEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
	conditionInitializedTrue := condition.True(api.CrdbInitializedCondition, conditions)
	conditionVersionCheckedTrue := condition.True(api.CrdbVersionChecked, conditions)

	// In order to do a partitioned update,
	// - the cluster should be initialized
	// - if the version validator is enabled, the version must be checked
	// - the current and desired versions should be non-empty and they must not match

	if !conditionInitializedTrue {
		return false
	}
	if featureVersionValidatorEnabled && !conditionVersionCheckedTrue {
		return false
	}

	versionWanted := cluster.GetVersionAnnotation()
	currentVersion := ss.Annotations[resource.CrdbVersionAnnotation]
	return currentVersion != versionWanted && currentVersion != "" && versionWanted != ""
}

func (cd *clusterDirector) needsPVCResize(cluster *resource.Cluster, ss *appsv1.StatefulSet) bool {
	conditions := cluster.Status().Conditions
	featureResizePVCEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.ResizePVC)
	conditionInitializedTrue := condition.True(api.CrdbInitializedCondition, conditions)

	// In order to resize PVCs,
	// - the resize PVC feature should be enabled
	// - the cluster must be initialized
	// - the data store's volume claim must not be nil and the stateful set must specify nonzero volume claim templates
	// - the size of the PVCs deployed must not match the size currently specified

	if !featureResizePVCEnabled {
		return false
	}
	if !conditionInitializedTrue {
		return false
	}

	if cluster.Spec().DataStore.VolumeClaim == nil || len(ss.Spec.VolumeClaimTemplates) == 0 {
		return false
	}
	stsStorageSizeDeployed := ss.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage()
	stsStorageSizeSet := cluster.Spec().DataStore.VolumeClaim.PersistentVolumeClaimSpec.Resources.Requests.Storage()
	return !stsStorageSizeDeployed.Equal(stsStorageSizeSet.DeepCopy())
}

func (cd *clusterDirector) needsDeploy(ctx context.Context, cluster *resource.Cluster, log logr.Logger) (bool, error) {
	conditions := cluster.Status().Conditions
	featureVersionValidatorEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
	conditionInitializedTrue := condition.True(api.CrdbInitializedCondition, conditions)
	conditionInitializedFalse := condition.False(api.CrdbInitializedCondition, conditions)
	conditionVersionCheckedTrue := condition.True(api.CrdbVersionChecked, conditions)

	// In order to deploy,
	// - the cluster initialized condition must be true or false (not unknown)
	// - if the version validator is enabled, the version must be checked
	// - at least one of the discovery service, public service, stateful set, and pod distribution budget specs must have
	//   changed in some way

	if !conditionInitializedTrue && !conditionInitializedFalse {
		return false, nil
	}
	if featureVersionValidatorEnabled && !conditionVersionCheckedTrue {
		return false, nil
	}

	r := resource.NewManagedKubeResource(ctx, cd.client, cluster, kube.AnnotatingPersister)

	kubernetesDistro, err := cd.kubeDistro.Get(ctx, cd.clientset, log)
	if err != nil {

		return false, err
	}

	labelSelector := r.Labels.Selector(cluster.Spec().AdditionalLabels)
	builders := []resource.Builder{
		resource.DiscoveryServiceBuilder{Cluster: cluster, Selector: labelSelector},
		resource.PublicServiceBuilder{Cluster: cluster, Selector: labelSelector},
		resource.StatefulSetBuilder{Cluster: cluster, Selector: labelSelector, Telemetry: kubernetesDistro},
		resource.PdbBuilder{Cluster: cluster, Selector: labelSelector},
	}

	for _, b := range builders {
		hasChanged, err := resource.Reconciler{
			ManagedResource: r,
			Builder:         b,
			Owner:           cluster.Unwrap(),
			Scheme:          cd.scheme,
		}.HasChanged()

		if err != nil {
			return false, err
		} else if hasChanged {
			return true, nil
		}
	}
	return false, nil
}

func (cd *clusterDirector) needsInitialization(cluster *resource.Cluster) bool {
	conditions := cluster.Status().Conditions
	featureVersionValidatorEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
	conditionInitializedFalse := condition.False(api.CrdbInitializedCondition, conditions)
	conditionVersionCheckedTrue := condition.True(api.CrdbVersionChecked, conditions)

	// In order to initialize,
	// - the cluster initialized condition must be false
	// - if the version validator is enabled, the version must be checked

	if !conditionInitializedFalse {
		return false
	}
	if featureVersionValidatorEnabled && !conditionVersionCheckedTrue {
		return false
	}
	return true
}

func (cd *clusterDirector) ActAtomically(actorCtx context.Context, cluster *resource.Cluster, a Actor, logger logr.Logger) error {
	clusterName := cluster.Name()
	clusterNamespace := cluster.Namespace()

	status := cluster.Status()
	currentState := status.DirectorState
	switch currentState {
	case DirectorStateAvailable:
		break
	case DirectorStateBusy:
		currentStateUpdatedAt := cluster.Status().DirectorStateUpdatedAt
		if currentStateUpdatedAt.Add(maxTimeBusy).Before(time.Now()) {
			return PermanentErr{Err: errors.New("director has timed out")}
		} else {
			return DirectorLockError{Err: errors.Newf("an actor is currently performing an action: %s", status.ActiveActor)}
		}
	default:
		return PermanentErr{Err: errors.Newf("director is in an unknown state: %s", currentState)}
	}

	lockCtx := context.Background()

	directorUpdatedAt, newGeneration := cluster.UpdateDirectorState(DirectorStateBusy, string(a.GetActionType()))
	if err := cd.client.Status().Update(lockCtx, cluster.Unwrap()); err != nil {
		return DirectorLockError{Err: fmt.Errorf("failed to acquire director lock: %w", err)}
	}

	actorErr := a.Act(actorCtx, cluster, logger)

	fetcher := resource.NewKubeFetcher(lockCtx, clusterNamespace, cd.client)
	for {
		cr := resource.ClusterPlaceholder(clusterName)
		if err := fetcher.Fetch(cr); err != nil {
			return PermanentErr{Err: fmt.Errorf("could not retrieve cluster: %w", err)}
		}
		refreshedCluster := resource.NewCluster(cr)
		status = refreshedCluster.Status()

		// It's possible that the newly retrieved cluster does not yet reflect the update just made. If this is the case,
		// try again.
		if status.DirectorObservedGeneration < newGeneration {
			time.Sleep(5 * time.Second)
			continue
		}

		if status.DirectorState != DirectorStateBusy || !status.DirectorStateUpdatedAt.Equal(&directorUpdatedAt) {
			return PermanentErr{Err: fmt.Errorf("active director lost lock; this should not have happened")}
		}

		refreshedCluster.UpdateDirectorState(DirectorStateAvailable, "")
		if err := cd.client.Status().Update(lockCtx, refreshedCluster.Unwrap()); err != nil {
			if !k8sErrors.IsConflict(err) {
				return PermanentErr{Err: fmt.Errorf("failed to set director back to available: %w", err)}
			}
			continue
		}
		break
	}

	return actorErr
}
