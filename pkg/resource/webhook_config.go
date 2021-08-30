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

	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	admv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
)

const (
	mutatingWebhookName     = "mcrdbcluster.kb.io"
	mutatingWebhookConfig   = "mutating-webhook-configuration"
	validatingWebhookName   = "vcrdbcluster.kb.io"
	validatingWebhookConfig = "validating-webhook-configuration"
)

// ErrWebhookNotFound is returned when the particular CRDB webhook is not defined.
type ErrWebhookNotFound struct {
	name string
}

func (e *ErrWebhookNotFound) Error() string {
	return fmt.Sprintf("webhook %s not found", e.name)
}

// ConfigureMutatingWebhook sets the CABundle for the mutating webhook's client config. This is used to secure
// communication between the K8S API server and our custom webhook server.
//
// caCert should be a PEM-encoded certificate (typically taken from the webhook secret's "tls.ca" value).
func ConfigureMutatingWebhook(ctx context.Context, api admv1.MutatingWebhookConfigurationInterface, caCert []byte) error {
	log := logr.FromContextOrDiscard(ctx).WithName("webhook_config").WithValues("resource", mutatingWebhookConfig)

	log.V(debugLevel).Info("Fetching mutating webhook configuration")
	config, err := api.Get(ctx, mutatingWebhookConfig, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to fetch webhook configuration")
		return errors.Wrap(err, "failed to fetch mutating webhook configuration")
	}

	idx := -1
	for i, wh := range config.Webhooks {
		if wh.Name == mutatingWebhookName {
			idx = i
			break
		}
	}

	if idx < 0 {
		err := &ErrWebhookNotFound{name: mutatingWebhookName}
		log.Error(err, "Failed to find webhook", "webhook", mutatingWebhookName)
		return errors.Wrap(err, "failed to find webhook")
	}

	config.Webhooks[idx].ClientConfig.CABundle = caCert
	log.V(debugLevel).Info("Updating webhook CA bundle", "webhook", mutatingWebhookName)
	_, err = api.Update(ctx, config, metav1.UpdateOptions{})
	return errors.Wrap(err, "failed to set CABundle for mutating webhook")
}

// ConfigureValidatingWebhook sets the CABundle for the validating webhook's client config. This is used to secure
// communication between the K8S API server and our custom webhook server.
//
// caCert should be a PEM-encoded certificate (typically taken from the webhook secret's "tls.ca" valuie).
func ConfigureValidatingWebhook(ctx context.Context, api admv1.ValidatingWebhookConfigurationInterface, caCert []byte) error {
	log := logr.FromContextOrDiscard(ctx).WithName("webhook_config").WithValues("resource", validatingWebhookConfig)

	log.V(debugLevel).Info("Fetching validating webhook configuration")
	config, err := api.Get(ctx, validatingWebhookConfig, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to fetch webhook configuration")
		return errors.Wrap(err, "failed to fetch validating webhook configuration")
	}

	idx := -1
	for i, wh := range config.Webhooks {
		if wh.Name == "vcrdbcluster.kb.io" {
			idx = i
			break
		}
	}

	if idx < 0 {
		err := &ErrWebhookNotFound{name: validatingWebhookName}
		log.Error(err, "Failed to find webhook", "webhook", validatingWebhookName)
		return errors.Wrap(err, "failed to find webhook")
	}

	config.Webhooks[idx].ClientConfig.CABundle = caCert
	log.V(debugLevel).Info("Updating webhook CA bundle", "webhook", mutatingWebhookName)
	_, err = api.Update(ctx, config, metav1.UpdateOptions{})
	return errors.Wrap(err, "failed to set CABundle for validating webhook")
}
