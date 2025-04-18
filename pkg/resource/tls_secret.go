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

package resource

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	caCrtKey = "ca.crt"
	caKey    = "ca.key"
)

// CreateTLSSecret returns a TLSSecreat struct that is
// used to store the certs via secrets.
func CreateTLSSecret(name string, r Resource) *TLSSecret {

	// TODO we are not saving a corev1.Secret that
	// is the tls secret type but is an opaque secret
	// we should use the correct type
	s := &TLSSecret{
		Resource: r,
		secret: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if s.secret.Data == nil {
		s.secret.Data = map[string][]byte{}
	}

	return s
}

func LoadTLSSecret(name string, r Resource) (*TLSSecret, error) {
	s := &TLSSecret{
		Resource: r,
		secret: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	err := s.Fetch(s.secret)

	s.secret = s.secret.DeepCopy()

	if s.secret.Data == nil {
		s.secret.Data = map[string][]byte{}
	}

	return s, err
}

type TLSSecret struct {
	Resource

	secret *corev1.Secret
}

func (s *TLSSecret) ReadyCA() bool {
	data := s.secret.Data

	if _, ok := data[caKey]; !ok {
		return false
	}

	return true
}

func (s *TLSSecret) Ready() bool {
	data := s.secret.Data
	if _, ok := data[caCrtKey]; !ok {
		return false
	}

	if _, ok := data[corev1.TLSCertKey]; !ok {
		return false
	}

	if _, ok := data[corev1.TLSPrivateKeyKey]; !ok {
		return false
	}

	return true
}

func (s *TLSSecret) UpdateKey(key []byte) error {
	newKey := append([]byte{}, key...)

	_, err := s.Persist(s.secret, func() error {
		s.secret.Data[corev1.TLSPrivateKeyKey] = newKey
		return nil
	})

	return err
}

func (s *TLSSecret) UpdateCertAndCA(cert, ca []byte, log logr.Logger) error {
	newCert, newCA := append([]byte{}, cert...), append([]byte{}, ca...)

	_, err := s.Persist(s.secret, func() error {
		s.secret.Data[corev1.TLSCertKey] = newCert
		s.secret.Data[caCrtKey] = newCA

		return nil
	})

	return err
}

// UpdateCertAndKeyAndCA updates three different certificates at the same time.
// It save the TLSCertKey, the CA, and the TLSPrivateKey in a secret.
func (s *TLSSecret) UpdateCertAndKeyAndCA(cert, key []byte, ca []byte, log logr.Logger) error {
	newCert, newCA := append([]byte{}, cert...), append([]byte{}, ca...)
	newKey := append([]byte{}, key...)

	_, err := s.Persist(s.secret, func() error {
		s.secret.Data[corev1.TLSCertKey] = newCert
		s.secret.Data[caCrtKey] = newCA
		s.secret.Data[corev1.TLSPrivateKeyKey] = newKey

		return nil
	})

	return err
}

// UpdateCAKey updates CA key
func (s *TLSSecret) UpdateCAKey(cakey []byte, log logr.Logger) error {
	newCAKey := append([]byte{}, cakey...)

	_, err := s.Persist(s.secret, func() error {
		s.secret.Data[caKey] = newCAKey
		return nil
	})

	return err
}

func (s *TLSSecret) CA() []byte {
	return s.secret.Data[caCrtKey]
}
func (s *TLSSecret) CAKey() []byte {
	return s.secret.Data[caKey]
}

func (s *TLSSecret) Key() []byte {
	return s.secret.Data[corev1.TLSCertKey]
}

func (s *TLSSecret) PriveKey() []byte {
	return s.secret.Data[corev1.TLSPrivateKeyKey]
}
