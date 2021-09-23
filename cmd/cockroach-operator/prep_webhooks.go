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

package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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

	// ensure the TLS secret exists
	secret, err := findOrCreateSecret(ctx, cs, ns)
	if err != nil {
		return errors.Wrap(err, "failed to find or create webhook TLS secret")
	}

	// write crt and key to certs dir
	if err := writeWebhookSecrets(secret, dir); err != nil {
		return err
	}

	// patch the hook config's CABundle
	return errors.Wrap(
		secret.ApplyWebhookConfig(ctx, cs.AdmissionregistrationV1()),
		"failed to patch the CABundle for the webhooks",
	)
}

func findOrCreateSecret(ctx context.Context, api kubernetes.Interface, ns string) (*resource.WebhookSecret, error) {
	secrets := api.CoreV1().Secrets(ns)
	secret, err := resource.LoadWebhookSecret(ctx, secrets)
	if err == nil {
		// secret already exists, use it.
		return secret, nil
	}

	if apiErrors.IsNotFound(err) {
		// secret needs to be created
		return resource.CreateWebhookSecret(ctx, secrets, ns)
	}

	// an unrecoverable error occured
	return nil, err
}

func writeWebhookSecrets(s *resource.WebhookSecret, dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create certs directory")
	}

	mode := os.FileMode(0600)

	if err := ioutil.WriteFile(filepath.Join(dir, "tls.crt"), s.Certificate(), mode); err != nil {
		return errors.Wrap(err, "failed to write TLS certificate")
	}

	return errors.Wrap(
		ioutil.WriteFile(filepath.Join(dir, "tls.key"), s.PrivateKey(), mode),
		"failed to write TLS private key",
	)
}
