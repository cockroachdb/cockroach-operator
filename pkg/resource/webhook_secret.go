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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	admv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	tlsCAKey = "tls.ca"

	webhookSecretKeySize  = 4096
	webhookSecretLifetime = 10 * 365 * 24 * time.Hour
	webhookSecretName     = "cockroach-operator-webhook-tls"
	webhookSecretOrg      = "Cockroach DB Operator"
	webhookServiceName    = "webhook-service"
)

// LoadWebhookSecret loads the secret from K8s and returns it.
func LoadWebhookSecret(ctx context.Context, client v1.SecretInterface) (*WebhookSecret, error) {
	log := logr.FromContextOrDiscard(ctx).WithName("webhook_secrets").WithValues("resource", webhookSecretName)
	log.V(2).Info("Fetching webhook secret")

	s, err := client.Get(ctx, webhookSecretName, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to fetch webhook secret")
		return nil, err
	}

	return webhookFromSecret(s), nil
}

// CreateWebhookSecret creates the cockroach-operator-webhook-tls secret. It will generate a new private key and
// certificate and store those values in the TLS secret.
func CreateWebhookSecret(ctx context.Context, client v1.SecretInterface) (*WebhookSecret, error) {
	log := logr.FromContextOrDiscard(ctx).WithName("webhook_secrets").WithValues("resource", webhookSecretName)
	log.V(2).Info("Creating webhook secret")

	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{webhookSecretOrg}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(webhookSecretLifetime),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	log.V(2).Info("Generating CA certificate")
	caCert, caPriv, err := generateCA(ca)
	if err != nil {
		log.Error(err, "Failed to generate CA certificate")
		return nil, err
	}

	log.V(2).Info("Generating server certificate")
	tlsCert, tlsKey, err := generateCert(ca, caPriv)
	if err != nil {
		log.Error(err, "Failed to generate server certificate")
		return nil, err
	}

	pemEncode := func(as string, raw []byte) []byte {
		return pem.EncodeToMemory(&pem.Block{Type: as, Bytes: raw})
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookSecretName,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			tlsCAKey:                pemEncode("CERTIFICATE", caCert),
			corev1.TLSCertKey:       pemEncode("CERTIFICATE", tlsCert),
			corev1.TLSPrivateKeyKey: pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(tlsKey.(*rsa.PrivateKey))),
		},
	}

	log.V(2).Info("Creating the secret resource")
	s, err := client.Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		log.Error(err, "Failed to create webhook secret")
		return nil, err
	}

	return webhookFromSecret(s), nil
}

func webhookFromSecret(s *corev1.Secret) *WebhookSecret {
	return &WebhookSecret{
		ca:   s.Data[tlsCAKey],
		key:  s.Data[corev1.TLSPrivateKeyKey],
		cert: s.Data[corev1.TLSCertKey],
	}
}

// WebhookSecret defines the TLS secret used to secure communication between the K8s API server and our webhooks.
type WebhookSecret struct {
	ca     []byte
	key    []byte
	cert   []byte
	client v1.SecretInterface
}

// PrivateKey returns the PEM-encoded private key.
func (ws *WebhookSecret) PrivateKey() []byte { return ws.key }

// CACertificate returns the PEM-encoded CA certificate.
func (ws *WebhookSecret) CACertificate() []byte { return ws.ca }

// Certificate returns the PEM-encoded certificate.
func (ws *WebhookSecret) Certificate() []byte { return ws.cert }

// ApplyWebhookConfig updates the CABundle for the webhook's ClientConfig. This is set for the mutating and
// validating hooks.
func (ws *WebhookSecret) ApplyWebhookConfig(ctx context.Context, api admv1.AdmissionregistrationV1Interface) error {
	if err := ConfigureMutatingWebhook(ctx, api.MutatingWebhookConfigurations(), ws.CACertificate()); err != nil {
		return err
	}

	return ConfigureValidatingWebhook(ctx, api.ValidatingWebhookConfigurations(), ws.CACertificate())
}

func generateCA(ca *x509.Certificate) ([]byte, crypto.PrivateKey, error) {
	pk, err := rsa.GenerateKey(rand.Reader, webhookSecretKeySize)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, &pk.PublicKey, pk)
	if err != nil {
		return nil, nil, err
	}

	return cert, pk, nil
}

func generateCert(ca *x509.Certificate, caPrivateKey crypto.PrivateKey) ([]byte, crypto.PrivateKey, error) {
	cert := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(webhookSecretLifetime),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		Subject: pkix.Name{
			CommonName:   webhookServiceName + ".default.svc",
			Organization: []string{webhookSecretOrg},
		},
		DNSNames: []string{
			webhookServiceName,
			webhookServiceName + ".default",
			webhookServiceName + ".default.svc",
			webhookServiceName + ".default.svc.cluster.local",
		},
	}

	pk, err := rsa.GenerateKey(rand.Reader, webhookSecretKeySize)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &pk.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, err
	}

	return certBytes, pk, nil
}
