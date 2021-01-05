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
	"strings"

	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newInitialize(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &initialize{
		action: newAction("initialize", scheme, cl),
		config: config,
	}
}

// initialize performs the initialization of the new cluster
type initialize struct {
	action

	config *rest.Config
}

func (init initialize) Handles(conds []api.ClusterCondition) bool {
	return condition.True(api.NotInitializedCondition, conds)
}

func (init initialize) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := init.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.Info("initializing CockroachDB")

	stsName := cluster.StatefulSetName()

	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      stsName,
	}
	ss := &appsv1.StatefulSet{}
	if err := init.client.Get(ctx, key, ss); err != nil {
		log.Info("failed to fetch statefulset")
		return kube.IgnoreNotFound(err)
	}

	status := &ss.Status

	if status.CurrentReplicas == 0 || status.CurrentReplicas < status.Replicas {
		log.Info("statefulset does not have all replicas up")
		return NotReadyErr{Err: errors.New("statefulset does not have all replicas up")}
	}

	cmd := []string{
		"/bin/bash",
		"-c",
		">- /cockroach/cockroach init " + cluster.SecureMode(),
	}

	_, stderr, err := kube.ExecInPod(init.scheme, init.config, cluster.Namespace(),
		fmt.Sprintf("%s-0", stsName), resource.DbContainerName, cmd)

	if err != nil && !alreadyInitialized(stderr) {
		// can happen if container has not finished its startup
		if strings.Contains(err.Error(), "unable to upgrade connection: container not found") ||
			strings.Contains(err.Error(), "does not have a host assigned") {
			return NotReadyErr{Err: errors.New("pod has not complitely started")}
		}

		return errors.Wrapf(err, "failed to initialize the cluster")
	}

	cluster.SetFalse(api.NotInitializedCondition)

	log.Info("completed")
	return nil
}

func alreadyInitialized(out string) bool {
	return strings.Contains(out, "cluster has already been initialized")
}
