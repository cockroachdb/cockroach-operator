// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package actor

import (
	"context"
	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Director interface {
	GetActorToExecute(context.Context, *resource.Cluster, logr.Logger) (Actor, error)
}

type clusterDirector struct {
	actors     map[api.ActionType]Actor
	client     client.Client
	scheme     *runtime.Scheme
	kubeDistro kube.KubernetesDistribution
	config     *rest.Config
}

func NewDirector(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Director {
	kd := kube.NewKubernetesDistribution()
	actors := map[api.ActionType]Actor{
		api.DecommissionAction:      newDecommission(scheme, cl, config),
		api.VersionCheckerAction:    newVersionChecker(scheme, cl, config),
		api.GenerateCertAction:      newGenerateCert(scheme, cl, config),
		api.PartitionedUpdateAction: newPartitionedUpdate(scheme, cl, config),
		api.ResizePVCAction:         newResizePVC(scheme, cl, config),
		api.DeployAction:            newDeploy(scheme, cl, config, kd),
		api.InitializeAction:        newInitialize(scheme, cl, config),
		api.ClusterRestartAction:    newClusterRestart(scheme, cl, config),
	}
	return &clusterDirector{
		actors:     actors,
		client:     cl,
		scheme:     scheme,
		kubeDistro: kd,
		config:     config,
	}
}

func (cd *clusterDirector) GetActorToExecute(ctx context.Context, cluster *resource.Cluster, log logr.Logger) (Actor, error) {
	if cd.needsRestart(cluster) {
		return cd.actors[api.ClusterRestartAction], nil
	}

	stsName := cluster.StatefulSetName()
	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      stsName,
	}
	ss := &appsv1.StatefulSet{}
	err := kube.IgnoreNotFound(cd.client.Get(ctx, key, ss))
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

	kubernetesDistro, err := cd.kubeDistro.Get(ctx, cd.config, log)
	if err != nil {
		return false, err
	}
	kubernetesDistro = "kubernetes-operator-" + kubernetesDistro

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
