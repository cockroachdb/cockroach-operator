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
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newClusterRestart(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &versionChecker{
		action: newAction("Crdb Version Validator", scheme, cl),
		config: config,
	}
}

// versionChecker performs the validation of the crdb image for the new cluster
type clusterRestart struct {
	action

	config *rest.Config
}

//GetActionType returns api.VersionCheckerAction action used to set the cluster status errors
func (v *clusterRestart) GetActionType() api.ActionType {
	return api.ClusterRestartAction
}

func (v *clusterRestart) Handles(conds []api.ClusterCondition) bool {
	return utilfeature.DefaultMutableFeatureGate.Enabled(features.ClusterRestart) && condition.True(api.InitializedCondition, conds) || condition.False(api.InitializedCondition, conds)
}

func (v *clusterRestart) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := v.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.V(int(zapcore.DebugLevel)).Info("starting cluster restart action")

	log.V(int(zapcore.DebugLevel)).Info("completed cluster restart")
	return nil
}
