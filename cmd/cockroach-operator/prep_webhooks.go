/*
Copyright 2026 The Cockroach Authors

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

package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/security"
	"github.com/cockroachdb/errors"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupWebhookTLS ensures that the webhook TLS secret exists, the necesary files are in place, and that the webhook
// client configuration has the correct CABundle for TLS. This should be called before starting the controller manager
// to ensure everything is in place at startup.
//
// Certificate rotation is as simple as deleting the pod and letting the deployment start a new one. If you're using
// your own certificates, be sure to update them before deleting the pod.
func SetupWebhookTLS(ctx context.Context, ns, dir string) error {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get REST config")
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to create client set")
	}

	webhookAPI := cs.AdmissionregistrationV1()
	secretsAPI := cs.CoreV1().Secrets(ns)

	// we create a new cert on each startup
	cert, err := resource.CreateWebhookCertificate(ctx, secretsAPI, ns)
	if err != nil {
		return errors.Wrap(err, "failed find or create webhook certificate")
	}

	// write them out
	if err := writeWebhookSecrets(cert, dir); err != nil {
		return errors.Wrap(err, "failed to write webhook certificate to disk")
	}

	ca, err := resource.FindOrCreateWebhookCA(ctx, secretsAPI)
	if err != nil {
		return errors.Wrap(err, "failed to find webhook CA certificate")
	}

	if err := resource.PatchMutatingWebhookConfig(ctx, webhookAPI.MutatingWebhookConfigurations(), ca); err != nil {
		return errors.Wrap(err, "failed to patch mutating webhook")
	}

	if err := resource.PatchValidatingWebhookConfig(ctx, webhookAPI.ValidatingWebhookConfigurations(), ca); err != nil {
		return errors.Wrap(err, "failed to patch validating webhook")
	}

	return nil
}

func writeWebhookSecrets(cert security.Certificate, dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create certs directory")
	}

	// r/w for current user only
	mode := os.FileMode(0600)

	if err := os.WriteFile(filepath.Join(dir, "tls.crt"), cert.Certificate(), mode); err != nil {
		return errors.Wrap(err, "failed to write TLS certificate")
	}

	return errors.Wrap(
		os.WriteFile(filepath.Join(dir, "tls.key"), cert.PrivateKey(), mode),
		"failed to write TLS private key",
	)
}
