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

package resource_test

import (
	"context"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/security"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFindOrCreateWebhookCA(t *testing.T) {
	ctx := context.Background()
	name := "cockroach-operator-webhook-ca"
	namespace := "bogus-ns"

	t.Run("creates CA when not found", func(t *testing.T) {
		api := fake.NewSimpleClientset().CoreV1().Secrets(namespace)

		cert, err := FindOrCreateWebhookCA(ctx, api)
		require.NoError(t, err)

		s, err := api.Get(ctx, name, metav1.GetOptions{})
		require.NoError(t, err)

		require.Equal(t, cert.Certificate(), s.Data[corev1.TLSCertKey])
		require.Equal(t, cert.PrivateKey(), s.Data[corev1.TLSPrivateKeyKey])
	})

	t.Run("uses exiting secret when it exists", func(t *testing.T) {
		api := fake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Type:       corev1.SecretTypeTLS,
			Data: map[string][]byte{
				corev1.TLSCertKey:       []byte("CERTIFICATE"),
				corev1.TLSPrivateKeyKey: []byte("PRIVATE KEY"),
			},
		}).CoreV1().Secrets(namespace)

		secret, err := api.Get(ctx, name, metav1.GetOptions{})
		require.NoError(t, err)

		crt, err := FindOrCreateWebhookCA(ctx, api)
		require.NoError(t, err)
		require.Equal(t, secret.Data[corev1.TLSCertKey], crt.Certificate())
		require.Equal(t, secret.Data[corev1.TLSPrivateKeyKey], crt.PrivateKey())
	})
}

func TestCreateWebhookCertificate(t *testing.T) {
	ctx := context.Background()
	namespace := "bogus-ns"
	dnsNames := []string{
		"webhook-service",
		"webhook-service.bogus-ns",
		"webhook-service.bogus-ns.svc",
		"webhook-service.bogus-ns.svc.cluster.local",
	}

	t.Run("generates CA when not found", func(t *testing.T) {
		api := fake.NewSimpleClientset().CoreV1().Secrets(namespace)

		cert, err := CreateWebhookCertificate(ctx, api, namespace)
		require.NoError(t, err)

		// will have been created by call to CreateWebhookCertificate above
		ca, err := FindOrCreateWebhookCA(ctx, api)
		require.NoError(t, err)

		crt, err := security.ParseCertificate(cert.Certificate())
		require.NoError(t, err)

		caCrt, err := security.ParseCertificate(ca.Certificate())
		require.NoError(t, err)

		require.Equal(t, caCrt.Subject.Organization, crt.Subject.Organization)
		require.Equal(t, dnsNames[0], crt.Subject.CommonName)
		require.Equal(t, dnsNames, crt.DNSNames)
	})

	t.Run("uses existing CA when it exists", func(t *testing.T) {
		api := fake.NewSimpleClientset().CoreV1().Secrets(namespace)

		// ensure CA already exists
		ca, err := FindOrCreateWebhookCA(ctx, api)
		require.NoError(t, err)

		cert, err := CreateWebhookCertificate(ctx, api, namespace)
		require.NoError(t, err)

		_, err = security.ParseCertificate(cert.Certificate())
		require.NoError(t, err)

		_, err = security.ParseCertificate(ca.Certificate())
		require.NoError(t, err)
	})
}

func TestPatchMutatingWebhookConfig(t *testing.T) {
	name := "mutating-webhook-configuration"

	tests := []struct {
		name   string
		config v1.MutatingWebhookConfiguration
		err    error
	}{
		{
			name: "valid definition found",
			config: v1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Webhooks: []v1.MutatingWebhook{
					{Name: "mcrdbcluster.kb.io"},
				},
			},
		},
		{
			name: "config not found",
			err:  &apiErrors.StatusError{},
		},
		{
			name: "webhook not defined",
			config: v1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: name},
			},
			err: &ErrWebhookNotFound{},
		},
	}

	crt, err := security.NewCACertificate()
	require.NoError(t, err)

	for _, tt := range tests {
		ctx := context.Background()
		api := fake.NewSimpleClientset().AdmissionregistrationV1().MutatingWebhookConfigurations()

		if tt.config.Name != "" {
			_, err := api.Create(ctx, &tt.config, metav1.CreateOptions{})
			require.NoError(t, err, tt.name)
		}

		err := PatchMutatingWebhookConfig(ctx, api, crt)
		if tt.err != nil {
			require.IsType(t, tt.err, errors.Cause(err), tt.name)
			continue
		}

		require.NoError(t, err, tt.name)

		cfg, err := api.Get(ctx, name, metav1.GetOptions{})
		require.NoError(t, err, tt.name)
		require.Equal(t, crt.Certificate(), cfg.Webhooks[0].ClientConfig.CABundle)
	}
}

func TestPatchValidatingWebhookConfig(t *testing.T) {
	name := "validating-webhook-configuration"

	tests := []struct {
		name   string
		config v1.ValidatingWebhookConfiguration
		err    error
	}{
		{
			name: "valid definition found",
			config: v1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Webhooks: []v1.ValidatingWebhook{
					{Name: "vcrdbcluster.kb.io"},
				},
			},
		},
		{
			name: "config not found",
			err:  &apiErrors.StatusError{},
		},
		{
			name: "webhook not defined",
			config: v1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: name},
			},
			err: &ErrWebhookNotFound{},
		},
	}

	crt, err := security.NewCACertificate()
	require.NoError(t, err)

	for _, tt := range tests {
		ctx := context.Background()
		api := fake.NewSimpleClientset().AdmissionregistrationV1().ValidatingWebhookConfigurations()

		if tt.config.Name != "" {
			_, err := api.Create(ctx, &tt.config, metav1.CreateOptions{})
			require.NoError(t, err, tt.name)
		}

		err := PatchValidatingWebhookConfig(ctx, api, crt)
		if tt.err != nil {
			require.IsType(t, tt.err, errors.Cause(err), tt.name)
			continue
		}

		require.NoError(t, err, tt.name)

		cfg, err := api.Get(ctx, name, metav1.GetOptions{})
		require.NoError(t, err, tt.name)
		require.Equal(t, crt.Certificate(), cfg.Webhooks[0].ClientConfig.CABundle)
	}
}
