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
	"fmt"

	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/tls"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	certs "k8s.io/api/certificates/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newRequestCert(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &requestCert{
		action: newAction("prepare_tls", scheme, cl),
		config: config,
	}
}

// requestCert issues node and root client certificates via Kubernetes cluster CA
type requestCert struct {
	action

	config *rest.Config
}

func (rc *requestCert) Handles(conds []api.ClusterCondition) bool {
	return condition.True(api.NotInitializedCondition, conds)
}

func (rc *requestCert) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := rc.log.WithValues("CrdbCluster", cluster.ObjectKey())

	if !cluster.Spec().TLSEnabled || cluster.Spec().NodeTLSSecret != "" {
		log.Info("Skipping TLS cert generation", "enabled", cluster.Spec().TLSEnabled, "secret", cluster.Spec().NodeTLSSecret)
		return nil
	}

	if err := rc.issueNodeCert(ctx, log, cluster); err != nil {
		return err
	}

	return rc.issueClientCert(ctx, log, cluster)
}

func (rc *requestCert) issueNodeCert(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.Info("requesting node certificate")

	secret, err := resource.LoadTLSSecret(cluster.NodeTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))
	if kube.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "failed to get node TLS secret")
	}

	if secret.Ready() {
		return nil
	}

	csrName := fmt.Sprintf("node.%s.%s.%s", cluster.Name(), cluster.Namespace(), cluster.Domain())

	usages := []certs.KeyUsage{
		certs.UsageDigitalSignature,
		certs.UsageKeyEncipherment,
		certs.UsageClientAuth,
		certs.UsageServerAuth,
	}

	request := tls.NewNodeCertificateRequest(cluster.PublicServiceName(),
		cluster.DiscoveryServiceName(),
		cluster.Domain(),
		cluster.Namespace())

	return rc.issue(ctx, csrName, request, secret, usages)
}

func (rc *requestCert) issueClientCert(ctx context.Context, log logr.Logger, cluster *resource.Cluster) error {
	log.Info("requesting client certificate")

	secret, err := resource.LoadTLSSecret(cluster.ClientTLSSecretName(),
		resource.NewKubeResource(ctx, rc.client, cluster.Namespace(), kube.DefaultPersister))
	if client.IgnoreNotFound(err) != nil {
		return errors.Wrap(err, "failed to get client TLS secret")
	}

	if secret.Ready() {
		return nil
	}

	csrName := fmt.Sprintf("root.%s.%s.%s", cluster.Name(), cluster.Namespace(), cluster.Domain())

	usages := []certs.KeyUsage{
		certs.UsageDigitalSignature,
		certs.UsageKeyEncipherment,
		certs.UsageClientAuth,
	}

	request := tls.NewClientCertificateRequest("root")

	return rc.issue(ctx, csrName, request, secret, usages)
}

func (rc *requestCert) issue(ctx context.Context, csrName string, request *x509.CertificateRequest,
	secret *resource.TLSSecret, usages []certs.KeyUsage) error {
	log := rc.log.WithValues("csr", csrName)

	log.Info("issuing certificate")

	csr, err := tls.InitCSR(ctx, rc.client, csrName)
	if err != nil {
		return errors.Wrapf(err, "failed to init CSR %s", csrName)
	}

	switch csr.Status {
	case tls.SigningRequestNotFound:
		log.Info("submitting a CSR")

		if err := tls.SignAndCreate(request, secret, csr.Unwrap(), usages); err != nil {
			return err
		}

		if err := rc.client.Create(ctx, csr.Unwrap()); err != nil {
			return errors.Wrapf(err, "failed to create CSR %s", csrName)
		}

		return NotReadyErr{Err: errors.New("client CSR is not ready, giving it some time")}
	case tls.SigningRequestPending:
		log.Info("approving CSR")
		if err := tls.Approve(ctx, rc.config, csr.Unwrap()); err != nil {
			return err
		}

		return NotReadyErr{Err: errors.New("client CSR is not ready, giving it some time")}
	case tls.SigningRequestApproved:
		log.Info("the CSR has been approved")

		ca, err := kube.GetClusterCA(ctx, rc.config)
		if err != nil {
			return errors.Wrap(err, "failed to fetch cluster CA certificate")
		}

		pemCert := csr.Unwrap().Status.Certificate
		if err := secret.UpdateCertAndCA(pemCert, ca, log); err != nil {
			return errors.Wrap(err, "failed to update client TLS secret certs")
		}

		return nil
	case tls.SigningRequestDenied:
		log.Info("request was denied")

		return PermanentErr{Err: errors.New("client CSR request was denied")}
	default:
		return errors.New("unknown CSR status")
	}
}
