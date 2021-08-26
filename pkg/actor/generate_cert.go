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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/security"
	"github.com/cockroachdb/cockroach-operator/pkg/util"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Options settable via command-line flags. See below for defaults.
var caCertificateLifetime time.Duration
var certificateLifetime time.Duration
var allowCAKeyReuse bool
var overwriteFiles bool
var generatePKCS8Key bool

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

// Act func generates the various certificates required and then stores
// the certificates in secrets.
func (rc *generateCert) Act(ctx context.Context, cluster *resource.Cluster) error {

	log := rc.log.WithValues("CrdbCluster", cluster.ObjectKey())

	if !cluster.Spec().TLSEnabled || cluster.Spec().NodeTLSSecret != "" {
		log.V(DEBUGLEVEL).Info("Skipping TLS cert generation", "enabled", cluster.Spec().TLSEnabled, "secret", cluster.Spec().NodeTLSSecret)
		return nil
	}

	// create the various temporary directories to store the certficates in
	// the directors will delete when the code is completed.
	certsDir, cleanup := util.CreateTempDir("certsDir")
	defer cleanup()
	rc.CertsDir = certsDir

	caDir, cleanupCADir := util.CreateTempDir("caDir")
	defer cleanupCADir()
	rc.CAKey = filepath.Join(caDir, "ca.key")

	// generate the base CA cert and key
	if err := rc.generateCA(ctx, log, cluster); err != nil {
		msg := "error generating CA"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	}
	var expirationDatePtr *string
	// generate the node certificate for the database to use
	if expirationDate, err := rc.generateNodeCert(ctx, log, cluster); err != nil {
		msg := "error generating Node Certificate"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	} else {
		expirationDatePtr = &expirationDate
	}

	// TODO if we save the node certificate but error on saving the client
	// certificate should we delete the node secret?

	// generate the client certificates for the database to use
	if err := rc.generateClientCert(ctx, log, cluster); err != nil {
		msg := "error generating Client Certificate"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	}

	// we force the saving of the status on the cluster and cancel the loop
	fetcher := resource.NewKubeFetcher(ctx, cluster.Namespace(), rc.client)
	newcr := resource.ClusterPlaceholder(cluster.Name())
	if err := fetcher.Fetch(newcr); err != nil {
		msg := "failed to retrieve CrdbCluster resource"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	}
	refreshedCluster := resource.NewCluster(newcr)
	refreshedCluster.SetAnnotationCertExpiration(*expirationDatePtr)
	refreshedCluster.SetTrue(api.CertificateGenerated)
	crdbobj := refreshedCluster.Unwrap()

	//save annotation first
	err := rc.client.Update(ctx, crdbobj)
	if err != nil && k8sErrors.IsConflict(err) {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {

			if err := fetcher.Fetch(newcr); err != nil {
				msg := "failed to retrieve CrdbCluster resource"
				log.Error(err, msg)
				return errors.Wrap(err, msg)
			}
			refreshedCluster := resource.NewCluster(newcr)
			refreshedCluster.SetAnnotationCertExpiration(*expirationDatePtr)
			refreshedCluster.SetTrue(api.CertificateGenerated)
			crdbobj := refreshedCluster.Unwrap()
			//save annotation first

			err = rc.client.Update(ctx, crdbobj)
			if err != nil {
				msg := "failed updating the annotations on request certificate will try again"
				log.Error(err, msg)
				return errors.Wrap(err, msg)
			}
			return err
		})
		if err != nil {
			msg := "failed saving the annotations on request certificate"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}
	} else if err != nil {
		msg := "failed saving the annotations on request certificate"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	}

	err = rc.client.Status().Update(ctx, crdbobj)
	// retrying if we have a conflict
	if err != nil && k8sErrors.IsConflict(err) {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := fetcher.Fetch(newcr); err != nil {
				msg := "failed to retrieve CrdbCluster resource"
				log.Error(err, msg)
				return errors.Wrap(err, msg)
			}
			refreshedCluster := resource.NewCluster(newcr)
			crdbobj := refreshedCluster.Unwrap()
			err = rc.client.Status().Update(ctx, crdbobj)
			if err != nil {
				msg := "failed saving the status on generate cert"
				log.Error(err, msg)
				return errors.Wrap(err, msg)
			}
			return err
		})
		if err != nil {
			msg := "failed saving the status on generate cert"
			log.Error(err, msg)
			return errors.Wrap(err, msg)
		}
	} else if err != nil {
		msg := "failed saving cluster status on generate cert"
		log.Error(err, msg)
		return errors.Wrap(err, msg)
	}

	return nil
}

func (rc *generateCert) generateCA(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.V(DEBUGLEVEL).Info("generating CA")
	// load the secret.  If it exists don't update the cert
	secret, err := resource.LoadTLSSecret(cluster.CASecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))

	if kube.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "failed to get ca key secret")
	}
	// if the secret is ready then don't update the secret
	// the Actor should have already generated the secret
	if secret.ReadyCA() {
		log.V(DEBUGLEVEL).Info("not updating ca key as it exists")
		return nil
	}

	err = errors.Wrap(
		security.CreateCAPair(
			rc.CertsDir,
			rc.CAKey,
			caCertificateLifetime,
			allowCAKeyReuse,
			overwriteFiles),
		"failed to generate CA cert and key")
	if err != nil {
		return err
	}
	// Read the ca key into memory
	cakey, err := ioutil.ReadFile(rc.CAKey)
	if err != nil {
		return errors.Wrap(err, "unable to read ca.key")
	}

	// create and save the TLS certificates into a secret
	secret = resource.CreateTLSSecret(cluster.CASecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))

	if err = secret.UpdateCAKey(cakey, log); err != nil {
		return errors.Wrap(err, "failed to update ca key secret ")
	}

	log.V(DEBUGLEVEL).Info("generated and saved ca key")
	return nil
}

// TODO we have an edge case that exists that the actor is not handling properly
// If any errors occurs and we have save secrets we may need to delete the secrets
// We can get into a race condition where the Node certifcate was created, but the Client certificate was not.
// Errors are thrown, and then the actor runs again.
// This time a new CA is generated, the Node secret is not updated, but the client certicate is generated
// using a new CA.

func (rc *generateCert) generateNodeCert(ctx context.Context, log logr.Logger, cluster *resource.Cluster) (string, error) {
	log.V(DEBUGLEVEL).Info("generating node certificate")

	// load the secret.  If it exists don't update the cert
	secret, err := resource.LoadTLSSecret(cluster.NodeTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))
	if kube.IgnoreNotFound(err) != nil {
		return "", errors.Wrap(err, "failed to get node TLS secret")
	}

	// if the secret is ready then don't update the secret
	// the Actor should have already generated the secret
	if secret.Ready() {
		log.V(DEBUGLEVEL).Info("not updating node certificate as it exists")
		return rc.getCertificateExpirationDate(ctx, log, secret.Key())
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
	err = errors.Wrap(
		security.CreateNodePair(
			rc.CertsDir,
			rc.CAKey,
			certificateLifetime,
			overwriteFiles,
			hosts),
		"failed to generate node certificate and key")

	if err != nil {
		return "", err
	}

	// Read the node certificates into memory
	ca, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "ca.crt"))
	if err != nil {
		return "", errors.Wrap(err, "unable to read ca.crt")
	}

	pemCert, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "node.crt"))
	if err != nil {
		return "", errors.Wrap(err, "unable to read node.crt")
	}

	pemKey, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "node.key"))
	if err != nil {
		return "", errors.Wrap(err, "unable to ready node.key")
	}

	// TODO we are not using the TLS secret type, but are using Opaque secrets.
	// We should refactor and use the TLS secret type

	// create and save the TLS certificates into a secret
	secret = resource.CreateTLSSecret(cluster.NodeTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))

	if err = secret.UpdateCertAndKeyAndCA(pemCert, pemKey, ca, log); err != nil {
		return "", errors.Wrap(err, "failed to update node TLS secret certs")
	}

	log.V(DEBUGLEVEL).Info("generated and saved node certificate and key")
	return rc.getCertificateExpirationDate(ctx, log, pemCert)
}

func (rc *generateCert) generateClientCert(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.V(DEBUGLEVEL).Info("generating client certificate")

	// load the secret.  If it exists don't update the cert
	secret, err := resource.LoadTLSSecret(cluster.ClientTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))
	if client.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "failed to get client TLS secret")
	}

	// if the secret is ready then don't update the secret
	// the Actor should have already generated the secret
	//but we should read the expiration date
	if secret.Ready() {
		log.V(DEBUGLEVEL).Info("not updating client certificate")
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
		return errors.Wrap(err, "unable to read ca.crt")
	}

	pemCert, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "client.root.crt"))
	if err != nil {
		return errors.Wrap(err, "unable to read client.root.crt")
	}

	pemKey, err := ioutil.ReadFile(filepath.Join(rc.CertsDir, "client.root.key"))
	if err != nil {
		return errors.Wrap(err, "unable to read client.root.key")
	}

	// create and save the TLS certificates into a secret
	secret = resource.CreateTLSSecret(cluster.ClientTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))

	if err = secret.UpdateCertAndKeyAndCA(pemCert, pemKey, ca, log); err != nil {
		return errors.Wrap(err, "failed to update client TLS secret certs")
	}

	log.V(DEBUGLEVEL).Info("generated and saved client certificate and key")
	return nil
}

func (rc *generateCert) getCertificateExpirationDate(ctx context.Context, log logr.Logger, pemCert []byte) (string, error) {
	log.V(DEBUGLEVEL).Info("getExpirationDate from cert")
	block, _ := pem.Decode(pemCert)
	if block == nil {
		return "", errors.New("failed to decode certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse certificate")
	}

	log.V(DEBUGLEVEL).Info("getExpirationDate from cert", "Not before:", cert.NotBefore.Format(time.RFC3339), "Not after:", cert.NotAfter.Format(time.RFC3339))
	return cert.NotAfter.Format(time.RFC3339), nil
}
