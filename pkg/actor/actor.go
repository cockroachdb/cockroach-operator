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
	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Different logging levels
var DEBUGLEVEL = int(zapcore.DebugLevel)
var WARNLEVEL = int(zapcore.WarnLevel)

//NotReadyErr strut
type NotReadyErr struct {
	Err error
}

func (e NotReadyErr) Error() string {
	return e.Err.Error()
}

//PermanentErr struct
type PermanentErr struct {
	Err error
}

func (e PermanentErr) Error() string {
	return e.Err.Error()
}

//InvalidContainerVersionError error used to stop requeue the request on failure
type InvalidContainerVersionError struct {
	Err error
}

func (e InvalidContainerVersionError) Error() string {
	return e.Err.Error()
}

//ValidationError error used to stop requeue the request on failure
type ValidationError struct {
	Err error
}

func (e ValidationError) Error() string {
	return e.Err.Error()
}

// Actor is one action against the cluster if the cluster resource state can be handled
type Actor interface {
	Act(context.Context, *resource.Cluster, logr.Logger) error
	GetActionType() api.ActionType
}

type Director interface {
	GetActorToExecute(context.Context, *resource.Cluster, logr.Logger) (Actor, error)
}

type clusterDirector struct {
	actors map[api.ActionType]Actor
	client client.Client
}

func NewDirector(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Director {
	actors := map[api.ActionType]Actor{
		api.DecommissionAction:      newDecommission(scheme, cl, config),
		api.VersionCheckerAction:    newVersionChecker(scheme, cl, config),
		api.GenerateCertAction:      newGenerateCert(scheme, cl, config),
		api.PartitionedUpdateAction: newPartitionedUpdate(scheme, cl, config),
		api.ResizePVCAction:         newResizePVC(scheme, cl, config),
		api.DeployAction:            newDeploy(scheme, cl, config, kube.NewKubernetesDistribution()),
		api.InitializeAction:        newInitialize(scheme, cl, config),
		api.ClusterRestartAction:    newClusterRestart(scheme, cl, config),
	}
	return &clusterDirector{
		actors: actors,
		client: cl,
	}
}

func (cd *clusterDirector) GetActorToExecute(ctx context.Context, cluster *resource.Cluster, log logr.Logger) (Actor, error) {
	conditions := cluster.Status().Conditions
	featureVersionValidatorEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
	featureDecommissionEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.Decommission)
	featureResizePVCEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.ResizePVC)
	featureClusterRestartEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.ClusterRestart)
	conditionInitializedTrue := condition.True(api.CrdbInitializedCondition, conditions)
	conditionInitializedFalse := condition.False(api.CrdbInitializedCondition, conditions)
	conditionVersionCheckedTrue := condition.True(api.CrdbVersionChecked, conditions)
	conditionVersionCheckedFalse := condition.False(api.CrdbVersionChecked, conditions)
	conditionCertificateGeneratedTrue := condition.True(api.CertificateGenerated, conditions)

	// cluster restart
	// decommission
	// version checker
	// generate cert
	// partitioned update
	// resize pvc
	// deploy
	// initialize

	if featureClusterRestartEnabled {
		if featureVersionValidatorEnabled && conditionVersionCheckedTrue && (conditionInitializedTrue || conditionInitializedFalse) {
			restartType := cluster.GetAnnotationRestartType()
			if restartType != "" {
				return cd.actors[api.ClusterRestartAction], nil
			}
		}
	}

	if featureDecommissionEnabled && conditionInitializedTrue {
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

		status := &ss.Status
		if status.CurrentReplicas == status.Replicas && status.CurrentReplicas > cluster.Spec().Nodes {
			return cd.actors[api.DecommissionAction], nil
		}
	}

	if featureVersionValidatorEnabled && conditionVersionCheckedFalse && (conditionInitializedTrue || conditionInitializedFalse) {
		return cd.actors[api.VersionCheckerAction], nil
	}

	if !conditionCertificateGeneratedTrue {
		return cd.actors[api.GenerateCertAction], nil
	}

	return nil, nil
}

func (cd *clusterDirector) GetActorsToExecute(cluster *resource.Cluster) []Actor {
	conditions := cluster.Status().Conditions
	featureVersionValidatorEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
	featureDecommissionEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.Decommission)
	featureResizePVCEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.ResizePVC)
	featureClusterRestartEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.ClusterRestart)
	conditionInitializedTrue := condition.True(api.CrdbInitializedCondition, conditions)
	conditionInitializedFalse := condition.False(api.CrdbInitializedCondition, conditions)
	conditionVersionCheckedTrue := condition.True(api.CrdbVersionChecked, conditions)
	conditionVersionCheckedFalse := condition.False(api.CrdbVersionChecked, conditions)
	conditionCertificateGeneratedTrue := condition.True(api.CertificateGenerated, conditions)

	var actorsToExecute []Actor

	if featureDecommissionEnabled && conditionInitializedTrue {
		actorsToExecute = append(actorsToExecute, cd.actors[api.DecommissionAction])
	}

	if featureVersionValidatorEnabled && conditionVersionCheckedFalse && (conditionInitializedTrue || conditionInitializedFalse) {
		actorsToExecute = append(actorsToExecute, cd.actors[api.VersionCheckerAction])
	}

	if !conditionCertificateGeneratedTrue {
		actorsToExecute = append(actorsToExecute, cd.actors[api.GenerateCertAction])
	}

	if featureVersionValidatorEnabled && conditionVersionCheckedTrue && conditionInitializedTrue {
		actorsToExecute = append(actorsToExecute, cd.actors[api.PartitionedUpdateAction])
	} else if !featureVersionValidatorEnabled && conditionInitializedTrue {
		actorsToExecute = append(actorsToExecute, cd.actors[api.PartitionedUpdateAction])
	}

	if featureResizePVCEnabled && conditionInitializedTrue {
		actorsToExecute = append(actorsToExecute, cd.actors[api.ResizePVCAction])
	}

	if featureVersionValidatorEnabled && conditionVersionCheckedTrue && (conditionInitializedTrue || conditionInitializedFalse) {
		actorsToExecute = append(actorsToExecute, cd.actors[api.DeployAction])
	} else if !featureVersionValidatorEnabled && (conditionInitializedTrue || conditionInitializedFalse) {
		actorsToExecute = append(actorsToExecute, cd.actors[api.DeployAction])
	}

	if featureVersionValidatorEnabled && conditionVersionCheckedTrue && conditionInitializedFalse {
		actorsToExecute = append(actorsToExecute, cd.actors[api.InitializeAction])
	} else if !featureVersionValidatorEnabled && conditionInitializedFalse {
		actorsToExecute = append(actorsToExecute, cd.actors[api.InitializeAction])
	}

	// TODO: conditionVersionCheckedTrue should probably be contingent on featureVersionValidatorEnabled, like with other actions
	if featureClusterRestartEnabled && conditionVersionCheckedTrue && (conditionInitializedTrue || conditionInitializedFalse) {
		actorsToExecute = append(actorsToExecute, cd.actors[api.ClusterRestartAction])
	}

	return actorsToExecute
}

func newAction(atype string, scheme *runtime.Scheme, cl client.Client) action {
	return action{
		client: cl,
		scheme: scheme,
	}
}

// action is the base set of common parameters required by other actions
type action struct {
	client client.Client
	scheme *runtime.Scheme
}
