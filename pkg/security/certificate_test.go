/*
Copyright 2024 The Cockroach Authors

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

package security_test

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/security"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func TestNewCACertificate(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		org     string
		keySize int
	}{
		{
			name:    "defaults",
			org:     "Self-Signed Issuer",
			keySize: 4096,
		},
		{
			name:    "custom org",
			opts:    []Option{OrgOption("My Org")},
			org:     "My Org",
			keySize: 4096,
		},
		{
			name:    "custom key size",
			opts:    []Option{KeySizeOption(2048)},
			org:     "Self-Signed Issuer",
			keySize: 2048,
		},
	}

	usage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	for _, tt := range tests {
		cert, err := NewCACertificate(tt.opts...)
		require.NoError(t, err, tt.name)
		require.NotNil(t, cert, tt.name)

		// validate certificate details
		ca, err := ParseCertificate(cert.Certificate())
		require.NoError(t, err, tt.name)
		require.Equal(t, []string{tt.org}, ca.Subject.Organization, tt.name)
		require.True(t, ca.IsCA, tt.name, tt.name)
		require.Equal(t, usage, ca.ExtKeyUsage, tt.name)
		require.Equal(t, x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign, ca.KeyUsage, tt.name)
		require.True(t, ca.BasicConstraintsValid, tt.name)

		pk, err := ParsePrivateKey(cert.PrivateKey())
		require.NoError(t, err, tt.name)
		require.Equal(t, tt.keySize, pk.N.BitLen(), tt.name)
	}
}

func TestNewCertificate(t *testing.T) {
	ca, err := NewCACertificate()
	require.NoError(t, err)

	tests := []struct {
		name     string
		opts     []Option
		org      string
		keySize  int
		dnsNames []string
	}{
		{
			name:    "defaults",
			org:     "Self-Signed Issuer",
			keySize: 4096,
		},
		{
			name:    "custom key size",
			opts:    []Option{KeySizeOption(2048)},
			org:     "Self-Signed Issuer",
			keySize: 2048,
		},
		{
			name:     "dns names",
			opts:     []Option{DNSNamesOption("svr1", "svr1.svc", "svr1.svc.cluster.local")},
			org:      "Self-Signed Issuer",
			keySize:  4096,
			dnsNames: []string{"svr1", "svr1.svc", "svr1.svc.cluster.local"},
		},
	}

	usage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	for _, tt := range tests {
		cert, err := NewCertificate(ca, tt.opts...)
		require.NoError(t, err, tt.name)
		require.NotNil(t, cert, tt.name)

		// validate certificate details
		crt, err := ParseCertificate(cert.Certificate())
		require.NoError(t, err, tt.name)
		require.Equal(t, []string{tt.org}, crt.Subject.Organization, tt.name)
		require.Equal(t, tt.dnsNames, crt.DNSNames, tt.name)

		if len(tt.dnsNames) > 0 {
			require.Equal(t, tt.dnsNames[0], crt.Subject.CommonName, tt.name)
		}

		require.False(t, crt.IsCA, tt.name, tt.name)
		require.Equal(t, usage, crt.ExtKeyUsage, tt.name)
		require.Equal(t, x509.KeyUsageDigitalSignature, crt.KeyUsage, tt.name)
		require.True(t, crt.BasicConstraintsValid, tt.name)

		pk, err := ParsePrivateKey(cert.PrivateKey())
		require.NoError(t, err, tt.name)
		require.Equal(t, tt.keySize, pk.N.BitLen(), tt.name)
	}
}

func TestParseCertificate(t *testing.T) {
	ca, err := NewCACertificate()
	require.NoError(t, err)

	tests := []struct {
		cert []byte
		err  string
	}{
		{cert: ca.Certificate()},
		{cert: []byte("-- nerp"), err: "failed to decode certificate"},
		{cert: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("nope")}), err: "failed to parse certificate"},
	}

	for _, tt := range tests {
		_, err := ParseCertificate(tt.cert)
		if tt.err != "" {
			require.Contains(t, err.Error(), tt.err)
			continue
		}

		require.NoError(t, err)
	}
}

func TestParsePrivateKey(t *testing.T) {
	ca, err := NewCACertificate()
	require.NoError(t, err)

	tests := []struct {
		pk  []byte
		err error
	}{
		{pk: ca.PrivateKey()},
		{pk: []byte("-- nerp"), err: ErrInvalidPEMBlock},
		{pk: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("nope")}), err: asn1.StructuralError{}},
	}

	for _, tt := range tests {
		_, err := ParsePrivateKey(tt.pk)
		require.IsType(t, tt.err, errors.Cause(err))
	}
}
