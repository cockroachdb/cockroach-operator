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

package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/cockroachdb/cockroach-operator/pkg/resource"
)

const rsaPrivateKeyPEMType = "RSA PRIVATE KEY"

func initPrivateKey(secret *resource.TLSSecret) (pemKey []byte, privateKey *rsa.PrivateKey, err error) {
	pemKey = secret.Key()

	// existing key is missing
	if pemKey == nil || len(pemKey) == 0 {
		return newPrivateKey()
	}

	// extract saved private key
	block, _ := pem.Decode(pemKey)
	if block == nil || block.Type != rsaPrivateKeyPEMType {
		return nil, nil, errors.New("failed to decode private key from secret")
	}

	if privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		return nil, nil, err
	}

	return pemKey, privateKey, nil
}

func newPrivateKey() (pemKey []byte, privateKey *rsa.PrivateKey, err error) {
	privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	pemKey = pem.EncodeToMemory(
		&pem.Block{
			Type:  rsaPrivateKeyPEMType,
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	return pemKey, privateKey, nil
}
