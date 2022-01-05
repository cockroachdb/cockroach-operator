/*
Copyright 2022 The Cockroach Authors

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

	"github.com/cockroachdb/cockroach-operator/pkg/security"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretsInterface is a subset of methods from client-go's v1.SecretsInterface
type SecretsInterface interface {
	Get(context.Context, string, metav1.GetOptions) (*corev1.Secret, error)
	Create(context.Context, *corev1.Secret, metav1.CreateOptions) (*corev1.Secret, error)
}

// CertificateFunc describes a function the generates Certificate objects.
type CertificateFunc func() (security.Certificate, error)

// FindOrCreateCertificateSecret gets the specified secret or creates it using the result of the creator func. The
// returned result, when not an error, is a secret of type SecretTypeTLS with the crt and key values set appropriately.
func FindOrCreateCertificateSecret(ctx context.Context, api SecretsInterface, name string, creator CertificateFunc) (security.Certificate, error) {
	log := logr.FromContextOrDiscard(ctx).WithValues("resource", name)

	s, err := api.Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return CertificateFromSecret(s), nil
	}

	if !apiErrors.IsNotFound(err) {
		log.Error(err, "Failed to lookup secret")
		return nil, errors.Wrap(err, "failed to lookup secret")
	}

	crt, err := creator()
	if err != nil {
		log.Error(err, "Failed to create certificate")
		return nil, errors.Wrap(err, "failed to create certificate")
	}

	s, err = api.Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       crt.Certificate(),
			corev1.TLSPrivateKeyKey: crt.PrivateKey(),
		},
	}, metav1.CreateOptions{})

	if err != nil {
		log.Error(err, "Failed to create TLS secret")
		return nil, errors.Wrap(err, "failed to create TLS secret")
	}

	return CertificateFromSecret(s), nil
}

// CertificateFromSecret creates a security.Certificate from the supplied secret. It is expected to be a TLS secret, or
// at least have the appropriate keys for Data entries.
func CertificateFromSecret(s *corev1.Secret) security.Certificate {
	return &certificate{
		crt: s.Data[corev1.TLSCertKey],
		pk:  s.Data[corev1.TLSPrivateKeyKey],
	}
}

type certificate struct {
	crt []byte
	pk  []byte
}

func (c *certificate) Certificate() []byte { return c.crt }
func (c *certificate) PrivateKey() []byte  { return c.pk }
