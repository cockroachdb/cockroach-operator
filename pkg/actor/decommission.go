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

	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/clustersql"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/database"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/scale"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newDecommission(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &decommission{
		action: newAction("decommission", scheme, cl),
		config: config,
	}
}

// decommission performs the initialization of the new cluster
type decommission struct {
	action

	config *rest.Config
}

func (d decommission) Handles(conds []api.ClusterCondition) bool {
	return condition.False(api.NotInitializedCondition, conds) && utilfeature.DefaultMutableFeatureGate.Enabled(features.Decommission)
}

func (d decommission) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := d.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.Info("check decommission oportunities")

	stsName := cluster.StatefulSetName()

	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      stsName,
	}
	ss := &appsv1.StatefulSet{}
	if err := d.client.Get(ctx, key, ss); err != nil {
		log.Info("decommission failed to fetch statefulset")
		return kube.IgnoreNotFound(err)
	}
	status := &ss.Status

	if status.CurrentReplicas == 0 || status.CurrentReplicas < status.Replicas {
		log.Info(" decommission statefulset does not have all replicas up")
		return NotReadyErr{Err: errors.New("decommission statefulset does not have all replicas up")}
	}

	nodes := uint(cluster.Spec().Nodes)
	//We do not scale down if the nodes field is less than 3
	//TODO @alina add validation webhook (see https://github.com/cockroachdb/cockroach-operator/issues/245)
	if nodes < 3 {
		log.Info("We cannot decommission if there are less than 3 nodes", "nodes", nodes)
		return errors.New("decommission with less than 3 nodes is not supported")
	}
	log.Info("replicas decommisioning", "status.CurrentReplicas", status.CurrentReplicas, "expected", cluster.Spec().Nodes)
	if status.CurrentReplicas <= cluster.Spec().Nodes {
		return nil
	}
	clientset, err := kubernetes.NewForConfig(d.config)
	if err != nil {
		return errors.Wrapf(err, "decommission failed to create kubernetes clientset")
	}
	// test to see if we are running inside of Kubernetes
	// If we are running inside of k8s we will not find this file.
	runningInsideK8s := inK8s("/var/run/secrets/kubernetes.io/serviceaccount/token")

	serviceName := cluster.PublicServiceName()
	if runningInsideK8s {
		log.Info("operator is running inside of kubernetes, connecting to service for db connection")
	} else {
		serviceName = fmt.Sprintf("%s-0.%s.%s", cluster.Name(), cluster.Name(), cluster.Namespace())
		log.Info("operator is NOT inside of kubernetes, connnecting to pod ordinal zero for db connection")
	}

	// The connection needs to use the discovery service name because of the
	// hostnames in the SSL certificates
	conn := &database.DBConnection{
		Ctx:              ctx,
		Client:           d.client,
		RestConfig:       d.config,
		ServiceName:      serviceName,
		Namespace:        cluster.Namespace(),
		DatabaseName:     "system", // TODO we need to use variable instead of string
		Port:             cluster.Spec().GRPCPort,
		RunningInsideK8s: runningInsideK8s,
	}

	// see https://github.com/cockroachdb/cockroach-operator/issues/204 for above TODO
	if cluster.Spec().TLSEnabled {
		conn.UseSSL = true
		conn.ClientCertificateSecretName = cluster.ClientTLSSecretName()
		conn.RootCertificateSecretName = cluster.NodeTLSSecretName()
	}
	db, err := database.NewDbConnection(conn)
	if err != nil {
		return errors.Wrapf(err, "failed to create database connection")
	}
	log.Info("opened db connection")
	defer db.Close()

	timeout, err := clustersql.RangeMoveDuration(ctx, db)
	if err != nil {
		return errors.Wrap(err, "failed to get range move duration")
	}

	drainer := scale.NewCockroachNodeDrainer(d.log, cluster.Namespace(), ss.Name, d.config, clientset, cluster.Spec().TLSEnabled, 3*timeout)
	pvcPruner := scale.PersistentVolumePruner{
		Namespace:   cluster.Namespace(),
		StatefulSet: ss.Name,
		ClientSet:   clientset,
		Logger:      d.log,
	}
	//we should start scale down
	scaler := scale.Scaler{
		Logger: d.log,
		CRDB: &scale.CockroachStatefulSet{
			ClientSet: clientset,
			Namespace: cluster.Namespace(),
			Name:      ss.Name,
		},
		Drainer:   drainer,
		PVCPruner: &pvcPruner,
	}
	if err := scaler.EnsureScale(ctx, nodes); err != nil {
		/// now check if the decommisiionStaleErr and update status
		log.Error(err, "decomission failed")
		cluster.SetFalse(api.DecommissionCondition)
		return nil
	}
	// TO DO @alina we will need to save the status foreach action
	cluster.SetTrue(api.DecommissionCondition)
	log.Info("decommission completed", "cond", ss.Status.Conditions)
	CancelLoop(ctx)
	return nil
}
