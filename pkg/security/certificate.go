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

package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/cockroachdb/errors"
)

// ErrInvalidPEMBlock is returned when parsing certificates or private keys fails.
var ErrInvalidPEMBlock = fmt.Errorf("invalid PEM block")

// options contains the configurable values for generating certificates.
type options struct {
	dnsNames  []string
	exp       time.Duration
	keySize   int
	org       string
	serialNum *big.Int
}

// Option defines a configuration option for certificate creation.
type Option interface {
	apply(*options)
}

type optionFn func(*options)

func (fn optionFn) apply(o *options) { fn(o) }

// DNSNamesOption sets the DNS names for the certificate. This option doesn't apply to CA certificates.
// Default: []
func DNSNamesOption(names ...string) Option {
	return optionFn(func(o *options) { o.dnsNames = names })
}

// ExpOption sets the valid duration for the certificate.
// Default: 10 years
func ExpOption(d time.Duration) Option {
	return optionFn(func(o *options) { o.exp = d })
}

// KeySizeOption sets the size of the private key.
// Default: 4096
func KeySizeOption(n int) Option {
	return optionFn(func(o *options) { o.keySize = n })
}

// OrgOption defines the issuing organization for the certificate.
// Default: Self-Signed Issuer
func OrgOption(org string) Option {
	return optionFn(func(o *options) { o.org = org })
}

// SerialOption sets the serial number for the certificate.
// Default: 1
func SerialOption(n *big.Int) Option {
	return optionFn(func(o *options) { o.serialNum = n })
}

// A Certificate consists of a private key and a certificate.
type Certificate interface {
	// Certificate returns the PEM-encoded x509 certificate.
	Certificate() []byte
	// PrivateKey returns the PEM-encoded PKCS1 private key.
	PrivateKey() []byte
}

type certificate struct {
	cert []byte
	pk   []byte
}

// Certificate returns the PEM-encoded certificate
func (c *certificate) Certificate() []byte { return c.cert }

// PrivateKey returns the PEM encoded private key
func (c *certificate) PrivateKey() []byte { return c.pk }

// NewCACertificate generates a new self-signed CA Certificate
func NewCACertificate(options ...Option) (Certificate, error) {
	opts := makeOptions(options)

	ca := &x509.Certificate{
		SerialNumber:          opts.serialNum,
		Subject:               pkix.Name{Organization: []string{opts.org}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(opts.exp),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	pk, err := rsa.GenerateKey(rand.Reader, opts.keySize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create private key")
	}

	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, &pk.PublicKey, pk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate")
	}

	return &certificate{
		cert: pemEncode("CERTIFICATE", cert),
		pk:   pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(pk)),
	}, nil
}

// NewCertificate generates a new Certificate using the supplied CA to sign it.
func NewCertificate(ca Certificate, options ...Option) (Certificate, error) {
	caCrt, err := ParseCertificate(ca.Certificate())
	if err != nil {
		return nil, err
	}

	caPk, err := ParsePrivateKey(ca.PrivateKey())
	if err != nil {
		return nil, err
	}

	opts := makeOptions(options)

	crt := &x509.Certificate{
		SerialNumber:          opts.serialNum,
		Subject:               caCrt.Subject,
		NotBefore:             time.Now(),
		NotAfter:              caCrt.NotAfter,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		DNSNames:              opts.dnsNames,
	}

	if len(opts.dnsNames) > 0 {
		crt.Subject.CommonName = opts.dnsNames[0]
	}

	pk, err := rsa.GenerateKey(rand.Reader, opts.keySize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create private key")
	}

	cert, err := x509.CreateCertificate(rand.Reader, crt, caCrt, &pk.PublicKey, caPk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create certificate")
	}

	return &certificate{
		cert: pemEncode("CERTIFICATE", cert),
		pk:   pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(pk)),
	}, nil
}

// ParseCertificate decodes the supplied bytes and parses an x509.Certificate from the result.
func ParseCertificate(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.Wrap(ErrInvalidPEMBlock, "failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	return cert, errors.Wrap(err, "failed to parse certificate")
}

// ParsePrivateKey decode the supplied bytes and parses an rsa.PrivateKey from the result.
func ParsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.Wrap(ErrInvalidPEMBlock, "failed to decode private key")
	}

	pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	return pk, errors.Wrap(err, "failed to parse private key")
}

func pemEncode(asType string, data []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: asType, Bytes: data})
}

func makeOptions(opts []Option) *options {
	o := &options{
		dnsNames:  []string{},
		exp:       10 * 365 * 24 * time.Hour,
		keySize:   4096,
		org:       "Self-Signed Issuer",
		serialNum: big.NewInt(1),
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}
