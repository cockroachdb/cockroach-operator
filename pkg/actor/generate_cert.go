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
	"io/ioutil"
	"path/filepath"
	"time"

	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/security"
	"github.com/cockroachdb/cockroach-operator/pkg/util"
	cr_errors "github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultKeySize = 2048

// We use 366 days on certificate lifetimes to at least match X years,
// otherwise leap years risk putting us just under.
const defaultCALifetime = 10 * 366 * 24 * time.Hour  // ten years
const defaultCertLifetime = 5 * 366 * 24 * time.Hour // five years

// Options settable via command-line flags. See below for defaults.
var keySize int
var caCertificateLifetime time.Duration
var certificateLifetime time.Duration
var allowCAKeyReuse bool
var overwriteFiles bool
var generatePKCS8Key bool

func initPreFlagsCertDefaults() {
	keySize = defaultKeySize
	caCertificateLifetime = defaultCALifetime
	certificateLifetime = defaultCertLifetime
	allowCAKeyReuse = false
	overwriteFiles = false
	generatePKCS8Key = false
}

func newGenerateCert(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {

	return &generateCert{
		action: newAction("generate_cert", scheme, cl),
		config: config,
	}
}

// generateCert issues node and root client certificates via Kubernetes cluster CA
type generateCert struct {
	action

	config   *rest.Config
	CertsDir string
	CAKey    string
}

//GetActionType returns api.RequestCertAction action used to set the cluster status errors
func (rc *generateCert) GetActionType() api.ActionType {
	return api.RequestCertAction
}

// Handles returns if this Actor can run.
func (rc *generateCert) Handles(conds []api.ClusterCondition) bool {
	return condition.True(api.NotInitializedCondition, conds)
}

// Act func generates the various certificates required and then stores
// the certificates in secrets.
func (rc *generateCert) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := rc.log.WithValues("CrdbCluster", cluster.ObjectKey())

	if !cluster.Spec().TLSEnabled || cluster.Spec().NodeTLSSecret != "" {
		log.Info("Skipping TLS cert generation", "enabled", cluster.Spec().TLSEnabled, "secret", cluster.Spec().NodeTLSSecret)
		return nil
	}

	// create the various temporary directories to store the certficates in
	// the directors will delete when the code is completed.
	certsDir, cleanup := util.CreateTempDir("certsDir")
	defer cleanup()
	rc.CertsDir = certsDir

	caDir, cleanupCADir := util.CreateTempDir("caDir")
	defer cleanupCADir()
	rc.CAKey = caDir + "/ca.key"

	// generate the base CA cert and key
	if err := rc.generateCA(ctx, log, cluster); err != nil {
		return err
	}

	// generate the node certificate for the database to use
	if err := rc.generateNodeCert(ctx, log, cluster); err != nil {
		return err
	}

	// generate the client certificates for the database to use
	return rc.generateClientCert(ctx, log, cluster)
}

func (rc *generateCert) generateCA(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.Info("generating CA")
	return cr_errors.Wrap(
		security.CreateCAPair(
			rc.CertsDir,
			rc.CAKey,
			keySize,
			caCertificateLifetime,
			allowCAKeyReuse,
			overwriteFiles),
		"failed to generate CA cert and key")
}

// TODO we have an edge case that exists that the actor is not handling properly
// If any errors occurs and we have save secrets we may need to delete the secrets
// We can get into a race condition where the Node certifcate was created, but the Client certificate was not.
// Errors are thrown, and then the actor runs again.
// This time a new CA is generated, the Node secret is not updated, but the client certicate is generated
// using a new CA.

func (rc *generateCert) generateNodeCert(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.Info("generating node certificate")

	// load the secret.  If it exists don't update the cert
	secret, err := resource.LoadTLSSecret(cluster.NodeTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))
	if kube.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "failed to get node TLS secret")
	}

	// if the secret is ready then don't update the secret
	// the Actor should have already generated the secret
	if secret.Ready() {
		return nil
	}

	// hosts are the various DNS names and IP address that have to exist in the Node certificates
	// for the database to function
	hosts := []string{
		"localhost",
		"127.0.0.1",
		cluster.PublicServiceName(),
		fmt.Sprintf("%s.%s", cluster.PublicServiceName(), cluster.Namespace()),
		fmt.Sprintf("%s.%s.%s", cluster.PublicServiceName(), cluster.Namespace(), cluster.Domain()),
		fmt.Sprintf("*.%s", cluster.DiscoveryServiceName()),
		fmt.Sprintf("*.%s.%s", cluster.DiscoveryServiceName(), cluster.Namespace()),
		fmt.Sprintf("*.%s.%s.%s", cluster.DiscoveryServiceName(), cluster.Namespace(), cluster.Domain()),
	}

	// create the Node Pair certificates
	err = cr_errors.Wrap(
		security.CreateNodePair(
			rc.CertsDir,
			rc.CAKey,
			keySize,
			certificateLifetime,
			overwriteFiles,
			hosts),
		"failed to generate node certificate and key")

	if err != nil {
		return err
	}

	// Read the node certificates into memory
	ca, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "ca.crt"))
	if err != nil {
		return err
	}

	pemCert, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "node.crt"))
	if err != nil {
		return err
	}

	pemKey, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "node.key"))
	if err != nil {
		return err
	}

	// TODO we are not using the TLS secret type, but are using Opaque secrets.
	// We should refactor and use the TLS secret type

	// create and save the TLS certificates into a secret
	secret = resource.CreateTLSSecret(cluster.NodeTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))

	if err = secret.UpdateCertAndKeyAndCA(pemCert, pemKey, ca, log); err != nil {
		return errors.Wrap(err, "failed to update node TLS secret certs")
	}

	log.Info("generated and saved node certificate and key")
	return nil
}

func (rc *generateCert) generateClientCert(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.Info("generating client certificate")

	// load the secret.  If it exists don't update the cert
	secret, err := resource.LoadTLSSecret(cluster.ClientTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))
	if client.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "failed to get client TLS secret")
	}

	// if the secret is ready then don't update the secret
	// the Actor should have already generated the secret
	if secret.Ready() {
		return nil
	}

	// Create the user for the certificate
	u := &security.SQLUsername{
		U: "root",
	}

	// Create the client certificates
	err = errors.Wrap(
		security.CreateClientPair(
			rc.CertsDir,
			rc.CAKey,
			keySize,
			certificateLifetime,
			overwriteFiles,
			*u,
			generatePKCS8Key),
		"failed to generate client certificate and key")
	if err != nil {
		return err
	}

	// Load the certificates into memory
	ca, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "ca.crt"))
	if err != nil {
		return err
	}

	pemCert, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "client.root.crt"))
	if err != nil {
		return err
	}

	pemKey, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "client.root.key"))
	if err != nil {
		return err
	}

	// create and save the TLS certificates into a secret
	secret = resource.CreateTLSSecret(cluster.ClientTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))

	if err = secret.UpdateCertAndKeyAndCA(pemCert, pemKey, ca, log); err != nil {
		return errors.Wrap(err, "failed to update client TLS secret certs")
	}

	log.Info("generated and saved client certificate and key")
	return nil
}
