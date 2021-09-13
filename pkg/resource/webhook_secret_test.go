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
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLoadWebhookSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("when secret exists", func(t *testing.T) {
		secrets := fake.NewSimpleClientset().CoreV1().Secrets("bogus-ns")
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cockroach-operator-webhook-tls",
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				"tls.ca":  []byte("CA_CERTIFICATE"),
				"tls.crt": []byte("CERTIFICATE"),
				"tls.key": []byte("PRIVATE KEY"),
			},
		}

		_, err := secrets.Create(ctx, secret, metav1.CreateOptions{})
		require.NoError(t, err)

		s, err := LoadWebhookSecret(ctx, secrets)
		require.NoError(t, err)
		require.Equal(t, "CA_CERTIFICATE", string(s.CACertificate()))
		require.Equal(t, "CERTIFICATE", string(s.Certificate()))
		require.Equal(t, "PRIVATE KEY", string(s.PrivateKey()))
	})

	t.Run("when not found", func(t *testing.T) {
		secrets := fake.NewSimpleClientset().CoreV1().Secrets("bogus-ns")

		s, err := LoadWebhookSecret(ctx, secrets)
		require.Nil(t, s)
		require.True(t, apiErrors.IsNotFound(err))
	})
}

func TestCreateWebhookSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("when successfully created", func(t *testing.T) {
		secrets := fake.NewSimpleClientset().CoreV1().Secrets("bogus-ns")

		s, err := CreateWebhookSecret(ctx, secrets)
		require.NoError(t, err)
		require.Contains(t, string(s.PrivateKey()), "BEGIN RSA PRIVATE KEY")
		require.Contains(t, string(s.Certificate()), "BEGIN CERTIFICATE")
		require.Contains(t, string(s.CACertificate()), "BEGIN CERTIFICATE")
	})

	t.Run("when it already exists", func(t *testing.T) {
		secrets := fake.NewSimpleClientset().CoreV1().Secrets("bogus-ns")

		_, err := CreateWebhookSecret(ctx, secrets)
		require.NoError(t, err)

		s, err := CreateWebhookSecret(ctx, secrets)
		require.Nil(t, s)
		require.True(t, apiErrors.IsAlreadyExists(err))
	})
}

func TestWebhookSecretApplyWebhookConfig(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                 string
		defineMutatingHook   bool
		defineValidatingHook bool
		err                  error
	}{
		{name: "both configs exist", defineMutatingHook: true, defineValidatingHook: true},
		{name: "missing mutating hook", defineValidatingHook: true, err: new(apiErrors.StatusError)},
		{name: "missing validating hook", defineMutatingHook: true, err: new(apiErrors.StatusError)},
	}

	for _, tt := range tests {
		client := fake.NewSimpleClientset()
		hookAPI := client.AdmissionregistrationV1()

		s, err := CreateWebhookSecret(ctx, client.CoreV1().Secrets("default"))
		require.NoError(t, err, tt.name)

		if tt.defineMutatingHook {
			_, err = hookAPI.
				MutatingWebhookConfigurations().
				Create(ctx, &v1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{Name: "mutating-webhook-configuration"},
					Webhooks:   []v1.MutatingWebhook{{Name: "mcrdbcluster.kb.io"}},
				}, metav1.CreateOptions{})

			require.NoError(t, err, tt.name)
		}

		if tt.defineValidatingHook {
			_, err = hookAPI.
				ValidatingWebhookConfigurations().
				Create(ctx, &v1.ValidatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{Name: "validating-webhook-configuration"},
					Webhooks:   []v1.ValidatingWebhook{{Name: "vcrdbcluster.kb.io"}},
				}, metav1.CreateOptions{})

			require.NoError(t, err, tt.name)
		}

		err = s.ApplyWebhookConfig(ctx, hookAPI)
		if tt.err != nil {
			require.IsType(t, tt.err, errors.Cause(err), tt.name)
			continue
		}

		require.NoError(t, err, tt.name)
	}
}
