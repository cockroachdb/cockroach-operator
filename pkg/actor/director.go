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
	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Director interface {
	GetActorsToExecute(*resource.Cluster) []Actor
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
		api.DecommissionAction:      newDecommission(cl, config, clientset),
		api.VersionCheckerAction:    newVersionChecker(scheme, cl, clientset),
		api.GenerateCertAction:      newGenerateCert(cl),
		api.PartitionedUpdateAction: newPartitionedUpdate(cl, config, clientset),
		api.ResizePVCAction:         newResizePVC(scheme, cl, clientset),
		api.DeployAction:            newDeploy(scheme, cl, kd, clientset),
		api.InitializeAction:        newInitialize(scheme, cl, config, clientset),
		api.ClusterRestartAction:    newClusterRestart(cl, config, clientset),
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
