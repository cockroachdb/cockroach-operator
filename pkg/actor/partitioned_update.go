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
	"os"
	"time"

	"github.com/cockroachdb/cockroach-operator/pkg/database"

	"github.com/Masterminds/semver/v3"
	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/update"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newPartitionedUpdate(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &partitionedUpdate{
		action: newAction("partitionedUpdate", scheme, cl),
		config: config,
	}
}

// upgrade handles minor and major version upgrades without finalization
type partitionedUpdate struct {
	action

	config *rest.Config
}

func (up *partitionedUpdate) Handles(conds []api.ClusterCondition) bool {
	return condition.False(api.NotInitializedCondition, conds)
}

// Act runs a new partitionUpdate.
// This update pattern handles the sql calls and workflow in order to
// update a cr cluster.  This is replacing the old update actor.
func (up *partitionedUpdate) Act(ctx context.Context, cluster *resource.Cluster) error {

	// TODO we have edge cases that we are not covering
	// see https://github.com/cockroachdb/cockroach-operator/issues/202

	log := up.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.Info("checking update opportunities, using a partitioned update")

	stsName := cluster.StatefulSetName()

	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      stsName,
	}
	statefulSet := &appsv1.StatefulSet{}
	if err := up.client.Get(ctx, key, statefulSet); err != nil {
		return errors.Wrap(err, "failed to fetch statefulset")
	}

	if statefulSetIsUpdating(statefulSet) {
		return NotReadyErr{Err: errors.New("statefulset is updating, waiting for the update to finish")}
	}

	// TODO we are relying on the container name for more than one purpose
	// it tells us the version and also we are finding it by name
	// We need to have it not tell us the version
	// See https://github.com/cockroachdb/cockroach-operator/issues/200

	dbContainer, err := kube.FindContainer(resource.DbContainerName, &statefulSet.Spec.Template.Spec)
	if err != nil {
		return err
	}

	// nothing to be done
	if dbContainer.Image == cluster.Spec().Image.Name {
		log.Info("no version changes needed")
		return nil
	}

	currentImageVersionString := getVersionFromImage(dbContainer.Image)
	if currentImageVersionString == "" {
		return fmt.Errorf("unknown CockroachDB version in container name: %s", dbContainer.Image)
	}

	currentVersion, err := semver.NewVersion(currentImageVersionString)
	if err != nil {
		return errors.Wrapf(err, "failed to parse container image version: %s", currentImageVersionString)
	}

	wantImageVersionString := getVersionFromImage(cluster.Spec().Image.Name)
	if wantImageVersionString == "" {
		return fmt.Errorf("unknown CockroachDB version in spec: %s", cluster.Spec().Image.Name)
	}

	wantVersion, err := semver.NewVersion(wantImageVersionString)
	if err != nil {
		return errors.Wrapf(err, "failed to parse spec image version: %s", wantImageVersionString)
	}

	wantContainer := getImageNameNoVersion(cluster.Spec().Image.Name)
	if wantContainer == "" {
		return fmt.Errorf("unknown CockroachDB container image %s", cluster.Spec().Image.Name)
	}

	// TODO we probably should make these items and more configurable
	// see https://github.com/cockroachdb/cockroach-operator/issues/203
	podUpdateTimeout := 10 * time.Minute
	podMaxPollingInterval := 30 * time.Minute
	sleeper := update.NewSleeper(1 * time.Minute)

	clientset, err := kubernetes.NewForConfig(up.config)
	if err != nil {
		return errors.Wrapf(err, "failed to create kubernetes clientset")
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
		Client:           up.client,
		RestConfig:       up.config,
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

	// TODO we may have an error case where the operator will not finish an update, but will
	// still try to make a database connection.
	// see https://github.com/cockroachdb/cockroach-operator/issues/205

	// Create a new database connection for the update.
	// TODO we may want to create this db connection later
	// see https://github.com/cockroachdb/cockroach-operator/issues/207

	db, err := database.NewDbConnection(conn)
	if err != nil {
		return errors.Wrapf(err, "failed to create database connection")
	}
	log.Info("opened db connection")
	defer db.Close()

	// TODO test downgrades
	// see https://github.com/cockroachdb/cockroach-operator/issues/208

	log.Info("update starting with partitioned update", "old version", currentImageVersionString, "new version", wantImageVersionString, "image", wantContainer)

	updateRoach := &update.UpdateRoach{
		CurrentVersion: currentVersion,
		WantVersion:    wantVersion,
		WantImageName:  wantContainer,
		StsName:        stsName,
		StsNamespace:   cluster.Namespace(),
		Db:             db,
	}

	k8sCluster := &update.UpdateCluster{
		Clientset:             clientset,
		PodUpdateTimeout:      podUpdateTimeout,
		PodMaxPollingInterval: podMaxPollingInterval,
		Sleeper:               sleeper,
	}

	err = update.UpdateClusterCockroachVersion(
		ctx,
		updateRoach,
		k8sCluster,
		log,
	)

	// TODO set status so that we will not try to update the cluster again
	// TODO set status to rollback cluster?
	// This work is pending the status field updates
	// see https://github.com/cockroachdb/cockroach-operator/issues/209

	if err != nil {
		return errors.Wrapf(err, "failed to update sts with partitioned update: %s", stsName)
	}

	// TODO set status that we are completed.
	log.Info("update completed with partitioned update", "new version", wantImageVersionString)
	CancelLoop(ctx)
	return nil
}

// inK8s checks to see if the a file exists
func inK8s(file string) bool {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
