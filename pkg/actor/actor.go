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
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
	Act(context.Context, *resource.Cluster) error
	GetActionType() api.ActionType
}

type Director interface {
	GetActorsToExecute(*resource.Cluster) []Actor
}

type clusterDirector struct {
	actors map[api.ActionType]Actor
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
	}
}

func (cd *clusterDirector) GetActorsToExecute(cluster *resource.Cluster) []Actor {
	conditions := cluster.Status().Conditions
	featureVersionValidatorEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator)
	featureDecommissionEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.Decommission)
	featureResizePVCEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.ResizePVC)
	featureClusterRestartEnabled := utilfeature.DefaultMutableFeatureGate.Enabled(features.ClusterRestart)
	conditionInitializedTrue := condition.True(api.InitializedCondition, conditions)
	conditionInitializedFalse := condition.False(api.InitializedCondition, conditions)
	conditionVersionCheckedTrue := condition.True(api.CrdbVersionChecked, conditions)
	conditionVersionCheckedFalse := condition.False(api.CrdbVersionChecked, conditions)

	var actorsToExecute []Actor

	if featureDecommissionEnabled && conditionInitializedTrue {
		actorsToExecute = append(actorsToExecute, cd.actors[api.DecommissionAction])
	}

	if featureVersionValidatorEnabled && conditionVersionCheckedFalse && (conditionInitializedTrue || conditionInitializedFalse) {
		actorsToExecute = append(actorsToExecute, cd.actors[api.VersionCheckerAction])
	}

	// TODO (this todo was copy/pasted from the deprecated Handles func): this is not working am I doing this correctly?
	// condition.True(api.CertificateGenerated, conds)
	if conditionInitializedFalse {
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

//Log var
var Log = logf.Log.WithName("action")

func newAction(atype string, scheme *runtime.Scheme, cl client.Client) action {
	return action{
		log:    Log.WithValues("action", atype),
		client: cl,
		scheme: scheme,
	}
}

// action is the base set of common parameters required by other actions
type action struct {
	log    logr.Logger
	client client.Client
	scheme *runtime.Scheme
}
