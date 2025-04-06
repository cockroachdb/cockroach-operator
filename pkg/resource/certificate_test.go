/*
Copyright 2025 The Cockroach Authors

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
	"errors"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/security"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCertificateFromSecret(t *testing.T) {
	secret := &corev1.Secret{
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("MyCA"),
			corev1.TLSPrivateKeyKey: []byte("MyKey"),
		},
	}

	cert := CertificateFromSecret(secret)
	require.Equal(t, cert.Certificate(), secret.Data[corev1.TLSCertKey])
	require.Equal(t, cert.PrivateKey(), secret.Data[corev1.TLSPrivateKeyKey])
}

func TestFindOrCreateCertificateSecret(t *testing.T) {
	ctx := context.Background()
	namespace := "bogus-ns"

	t.Run("uses existing values when found", func(t *testing.T) {
		api := fake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "sshhh", Namespace: namespace},
			Type:       corev1.SecretTypeTLS,
			Data: map[string][]byte{
				corev1.TLSCertKey:       []byte("CERTIFICATE"),
				corev1.TLSPrivateKeyKey: []byte("PRIVATE KEY"),
			},
		}).CoreV1().Secrets(namespace)

		cert, err := FindOrCreateCertificateSecret(ctx, api, "sshhh", nil)
		require.NoError(t, err)
		require.Equal(t, "CERTIFICATE", string(cert.Certificate()))
		require.Equal(t, "PRIVATE KEY", string(cert.PrivateKey()))
	})

	t.Run("creates new secret when not found", func(t *testing.T) {
		api := fake.NewSimpleClientset().CoreV1().Secrets(namespace)

		cert, err := FindOrCreateCertificateSecret(ctx, api, "sshhh", func() (security.Certificate, error) {
			return security.NewCACertificate()
		})

		require.NoError(t, err)

		secret, err := api.Get(ctx, "sshhh", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, cert.Certificate(), secret.Data[corev1.TLSCertKey])
		require.Equal(t, cert.PrivateKey(), secret.Data[corev1.TLSPrivateKeyKey])
	})

	t.Run("propagates error from creator func", func(t *testing.T) {
		api := fake.NewSimpleClientset().CoreV1().Secrets(namespace)

		cert, err := FindOrCreateCertificateSecret(ctx, api, "sshhh", func() (security.Certificate, error) {
			return nil, errors.New("boom")
		})

		require.Nil(t, cert)
		require.EqualError(t, err, "failed to create certificate: boom")

		s, err := api.Get(ctx, "sshhh", metav1.GetOptions{})
		require.Nil(t, s)
		require.True(t, apiErrors.IsNotFound(err))
	})
}
