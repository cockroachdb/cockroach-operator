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

package resource

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach-operator/pkg/security"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	mutatingHookConfig   = "cockroach-operator-mutating-webhook-configuration"
	mutatingHookName     = "mcrdbcluster.kb.io"
	validatingHookConfig = "cockroach-operator-validating-webhook-configuration"
	validatingHookName   = "vcrdbcluster.kb.io"
	webhookCASecret      = "cockroach-operator-webhook-ca"
	webhookSecretOrg     = "Cockroach DB Operator"
	webhookService       = "cockroach-operator-webhook-service"
)

// ErrWebhookNotFound is returned when the particular CRDB webhook is not defined.
type ErrWebhookNotFound struct {
	name string
}

func (e *ErrWebhookNotFound) Error() string {
	return fmt.Sprintf("webhook %s not found", e.name)
}

// MutatingWebhookConfigInterface is a subset of client-go's MutatingWebhookConfigurationInterface
type MutatingWebhookConfigInterface interface {
	Get(context.Context, string, metav1.GetOptions) (*v1.MutatingWebhookConfiguration, error)
	Update(context.Context, *v1.MutatingWebhookConfiguration, metav1.UpdateOptions) (*v1.MutatingWebhookConfiguration, error)
}

// ValidatingWebhookConfigInterface is a subset of client-go's MutatingWebhookConfigurationInterface
type ValidatingWebhookConfigInterface interface {
	Get(context.Context, string, metav1.GetOptions) (*v1.ValidatingWebhookConfiguration, error)
	Update(context.Context, *v1.ValidatingWebhookConfiguration, metav1.UpdateOptions) (*v1.ValidatingWebhookConfiguration, error)
}

// FindOrCreateWebhookCA ensures the webhook CA secret exists, creating it if it's not found. This certificate is used
// to sign the webhook server certificate which allows secure communication with the K8s API.
//
// If you'd like to use your own CA, create the cockroach-operator-webhook-ca TLS secret in the target namespace before
// starting the controller manager.
func FindOrCreateWebhookCA(ctx context.Context, api SecretsInterface) (security.Certificate, error) {
	d := int(zapcore.DebugLevel)
	log := logr.FromContextOrDiscard(ctx).WithName("webhook-setup")

	ca, err := FindOrCreateCertificateSecret(ctx, api, webhookCASecret, func() (security.Certificate, error) {
		log.V(d).Info("Creating a new CA certificate")
		return security.NewCACertificate(security.OrgOption(webhookSecretOrg))
	})

	if err != nil {
		log.Error(err, "Failed to get webhook CA certificate")
	}

	return ca, errors.Wrap(err, "failed to get webhook CA certificate")
}

// CreateWebhookCertificate generates a new server certificate signed with the webhook CA cert. This certificate is not
// saved anywhere in K8s. The idea is that rotation of the certificate can be performed simply by rolling the manager
// deployment.
func CreateWebhookCertificate(ctx context.Context, api SecretsInterface, ns string) (security.Certificate, error) {
	log := logr.FromContextOrDiscard(ctx).WithName("webhook-setup")

	ca, err := FindOrCreateWebhookCA(ctx, api)
	if err != nil {
		// logging and wrapping already done in FindOrCreateWebhookCA
		return nil, err
	}

	log.Info("Generating webhook certificate")
	crt, err := security.NewCertificate(ca, security.DNSNamesOption(
		webhookService,
		fmt.Sprintf("%s.%s", webhookService, ns),
		fmt.Sprintf("%s.%s.svc", webhookService, ns),
		fmt.Sprintf("%s.%s.svc.cluster.local", webhookService, ns),
	))

	if err != nil {
		log.Error(err, "Failed to create webhook certificate")
	}

	return crt, errors.Wrap(err, "failed to create webhook certificate")
}

// PatchMutatingWebhookConfig updates the mutating webhook's client config such the CABundle is set to the supplied
// certificate's Certificate value. This is necessary for secure communication between the webhook and the K8s API
// server.
func PatchMutatingWebhookConfig(ctx context.Context, api MutatingWebhookConfigInterface, cert security.Certificate) error {
	log := logr.
		FromContextOrDiscard(ctx).
		WithName("webhook-setup").
		WithValues("config", mutatingHookConfig, "webhook", mutatingHookName)

	log.Info("Patching CABundle for mutating webhook")

	config, err := api.Get(ctx, mutatingHookConfig, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to find mutating webhook configuration")
		return errors.Wrap(err, "failed to find mutating webhook configuration")
	}

	idx := -1
	for i, w := range config.Webhooks {
		if w.Name == mutatingHookName {
			idx = i
			break
		}
	}

	if idx < 0 {
		err := &ErrWebhookNotFound{name: mutatingHookName}
		log.Error(err, "Failed to find webhook")
		return errors.Wrap(err, "failed to find webhook")
	}

	config.Webhooks[idx].ClientConfig.CABundle = cert.Certificate()

	if _, err = api.Update(ctx, config, metav1.UpdateOptions{}); err != nil {
		log.Error(err, "Failed to set CABundle for mutating webhook")
	}

	return errors.Wrap(err, "failed to set CABundle for mutating webhook")
}

// PatchValidatingWebhookConfig updates the mutating webhook's client config such the CABundle is set to the supplied
// certificate's Certificate value. This is necessary for secure communication between the webhook and the K8s API
// server.
func PatchValidatingWebhookConfig(ctx context.Context, api ValidatingWebhookConfigInterface, cert security.Certificate) error {
	log := logr.
		FromContextOrDiscard(ctx).
		WithName("webhook-setup").
		WithValues("config", mutatingHookConfig, "webhook", mutatingHookName)

	log.Info("Patching CABundle for validating webhook")

	config, err := api.Get(ctx, validatingHookConfig, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to find validating webhook configuration")
		return errors.Wrap(err, "failed to find validating webhook configuration")
	}

	idx := -1
	for i, w := range config.Webhooks {
		if w.Name == validatingHookName {
			idx = i
			break
		}
	}

	if idx < 0 {
		err := &ErrWebhookNotFound{name: validatingHookName}
		log.Error(err, "Failed to find webhook")
		return errors.Wrap(err, "failed to find webhook")
	}

	config.Webhooks[idx].ClientConfig.CABundle = cert.Certificate()

	if _, err = api.Update(ctx, config, metav1.UpdateOptions{}); err != nil {
		log.Error(err, "Failed to set CABundle for validating webhook")
	}

	return errors.Wrap(err, "failed to set CABundle for validating webhook")
}
