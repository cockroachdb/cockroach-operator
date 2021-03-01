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
	"os"
	"time"

	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
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

func (rc *generateCert) Handles(conds []api.ClusterCondition) bool {
	return condition.True(api.NotInitializedCondition, conds)
}

func (rc *generateCert) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := rc.log.WithValues("CrdbCluster", cluster.ObjectKey())

	if !cluster.Spec().TLSEnabled || cluster.Spec().NodeTLSSecret != "" {
		log.Info("Skipping TLS cert generation", "enabled", cluster.Spec().TLSEnabled, "secret", cluster.Spec().NodeTLSSecret)
		return nil
	}

	// Create temp directories to store the certificates
	certsDir, err := ioutil.TempDir("", "certsDir")

	if err != nil {
		return err
	}

	rc.CertsDir = certsDir

	defer os.RemoveAll(certsDir)

	caDir, err := ioutil.TempDir("", "caDir")

	if err != nil {
		return err
	}

	rc.CAKey = caDir + "/ca.key"

	defer os.RemoveAll(caDir)

	if err := rc.generateCA(ctx, log, cluster); err != nil {
		return err
	}

	if err := rc.generateNodeCert(ctx, log, cluster); err != nil {
		return err
	}

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

func (rc *generateCert) generateNodeCert(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.Info("generating node certificate")

	secret, err := resource.LoadTLSSecret(cluster.NodeTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))
	if kube.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "failed to get node TLS secret")
	}

	if secret.Ready() {
		return nil
	}

	csrName := fmt.Sprintf("node.%s.%s.%s", cluster.Name(), cluster.Namespace(), cluster.Domain())
	args := []string{csrName}

	err = cr_errors.Wrap(
		security.CreateNodePair(
			rc.CertsDir,
			rc.CAKey,
			keySize,
			certificateLifetime,
			overwriteFiles,
			args),
		"failed to generate node certificate and key")

	if err != nil {
		return err
	}

	ca, err := ioutil.ReadFile(rc.CAKey)
	if err != nil {
		return err
	}

	pemCert, err := ioutil.ReadFile(rc.CertsDir + "/node.crt")
	if err != nil {
		return err
	}

	if err = secret.UpdateCertAndCA(pemCert, ca, log); err != nil {
		return errors.Wrap(err, "failed to update client TLS secret certs")
	}

	return nil
}

func (rc *generateCert) generateClientCert(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.Info("generating client certificate")

	secret, err := resource.LoadTLSSecret(cluster.ClientTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))
	if client.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "failed to get client TLS secret")
	}

	if secret.Ready() {
		return nil
	}

	if username, err = sql.NormalizeAndValidateUsernameNoBlocklist("root"); err != nil {
		return errors.Wrap(err, "failed to generate client certificate and key")
	}

	err = errors.Wrap(
		security.CreateClientPair(
			rc.CertsDir,
			rc.CAKey,
			keySize,
			certificateLifetime,
			overwriteFiles,
			"root",
			generatePKCS8Key),
		"failed to generate client certificate and key")
	if err != nil {
		return err
	}
	ca, err := ioutil.ReadFile(rc.CAKey)
	if err != nil {
		return err
	}

	pemCert, err := ioutil.ReadFile(rc.CertsDir + "/client.root.crt")
	if err != nil {
		return err
	}

	if err = secret.UpdateCertAndCA(pemCert, ca, log); err != nil {
		return errors.Wrap(err, "failed to update client TLS secret certs")
	}

	return nil
}
