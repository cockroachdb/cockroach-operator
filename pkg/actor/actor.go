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

// NewOperatorActions creates a slice of actors that control the actions or actors for the operator.
// The order of the slice is critical so that the actors run in order, for instance update has to
// happen before deploy.
func NewOperatorActions(scheme *runtime.Scheme, cl client.Client, config *rest.Config) map[api.ActionType]Actor {

	// The order of these actors MATTERS.
	// We need to have update before deploy so that
	// updates run before the deploy actor, or
	// deploy will update the STS container and not deploy.
	// Decommission needs to be first, it is not dependant on versionchecker.

	// Actors that controlled by featuregates
	// have the featuregate check above or in there handles func.
	return map[api.ActionType]Actor{
		api.DecommissionAction:   newDecommission(scheme, cl, config),
		api.VersionCheckerAction: newVersionChecker(scheme, cl, config),
		api.GenerateCertAction:   newGenerateCert(scheme, cl, config),
		api.PartialUpdateAction:  newPartitionedUpdate(scheme, cl, config),
		api.ResizePVCAction:      newResizePVC(scheme, cl, config),
		api.DeployAction:         newDeploy(scheme, cl, config, kube.NewKubernetesDistribution()),
		api.InitializeAction:     newInitialize(scheme, cl, config),
		api.ClusterRestartAction: newClusterRestart(scheme, cl, config),
	}
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
