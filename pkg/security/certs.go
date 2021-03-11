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

package security

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cockroachdb/errors"
)

// This code contains the funcs used in https://github.com/cockroachdb/cockroach/blob/19951d2ad7a8eb3c20c38c6e55e1414549d1850f/pkg/security/certs.go
// We want to replace this code using a libary that does not exist yet, for now we are going to use the CRDB binary
// directly

// Instead of using custom code to generate the certificates this code executes the crdb binary which then generates the certificates

// SQLUsername is used to define the username created in the client certificate
type SQLUsername struct {
	U string
}

const (
	keyFileMode  = 0600
	certFileMode = 0644
)

// PemUsage indicates the purpose of a given certificate.
type PemUsage uint32

const (
	_ PemUsage = iota
	// CAPem describes the main CA certificate.
	CAPem
	// TenantClientCAPem describes the CA certificate used to broker authN/Z for SQL
	// tenants wishing to access the KV layer.
	TenantClientCAPem
	// ClientCAPem describes the CA certificate used to verify client certificates.
	ClientCAPem
	// UICAPem describes the CA certificate used to verify the Admin UI server certificate.
	UICAPem
	// NodePem describes the server certificate for the node, possibly a combined server/client
	// certificate for user Node if a separate 'client.node.crt' is not present.
	NodePem
	// UIPem describes the server certificate for the admin UI.
	UIPem
	// ClientPem describes a client certificate.
	ClientPem
	// TenantClientPem describes a SQL tenant client certificate.
	TenantClientPem

	// Maximum allowable permissions.
	maxKeyPermissions os.FileMode = 0700
	// Filename extenstions.
	certExtension = `.crt`
	keyExtension  = `.key`
	// Certificate directory permissions.
	defaultCertsDirPerm = 0700
)

// The following constants are used to run the crdb binary
const (
	CR            string = "cockroach"
	CERT          string = "cert"
	CREATE_CA     string = "create-ca"
	CREATE_NODE   string = "create-node"
	CREATE_CLIENT string = "create-client"

	CERTS_DIR string = "--certs-dir=%s"
	CA_KEY    string = "--ca-key=%s"
)

// CreateCAPair creates a general CA certificate and associated key.
func CreateCAPair(
	certsDir, caKeyPath string,
	keySize int,
	lifetime time.Duration,
	allowKeyReuse bool,
	overwrite bool,
) error {
	return createCACertAndKey(certsDir, caKeyPath, CAPem, keySize, lifetime, allowKeyReuse, overwrite)
}

// This again is a copy of code used in the crdb hence why we are using the switch statement

// createCACertAndKey creates a CA key and a CA certificate.
// If the certs directory does not exist, it is created.
// If the key does not exist, it is created.
// The certificate is written to the certs directory. If the file already exists,
// we append the original certificates to the new certificate.
//
// The filename of the certificate file must be specified.
// It should be one of:
// - ca.crt: the general CA certificate
// - ca-client.crt: the CA certificate to verify client certificates
func createCACertAndKey(certsDir, caKeyPath string, caType PemUsage, keySize int, lifetime time.Duration, allowKeyReuse bool, overwrite bool) error {
	if len(caKeyPath) == 0 {
		return errors.New("the path to the CA key is required")
	}
	if len(certsDir) == 0 {
		return errors.New("the path to the certs directory is required")
	}
	if caType != CAPem &&
		caType != TenantClientCAPem &&
		caType != ClientCAPem &&
		caType != UICAPem {

		return fmt.Errorf("caType argument to createCACertAndKey must be one of CAPem (%d), ClientCAPem (%d), or UICAPem (%d), got: %d", CAPem, ClientCAPem, UICAPem, caType)
	}

	certsDirParam := fmt.Sprintf(CERTS_DIR, certsDir)
	caKeyParam := fmt.Sprintf(CA_KEY, caKeyPath)

	switch caType {
	case CAPem:
		// run the crdb binary to generate the CA
		execCmd(CREATE_CA, certsDirParam, caKeyParam)
	case TenantClientCAPem:
		return errors.Newf("unknown CA type %v", caType)
	case ClientCAPem:
		return errors.Newf("unknown CA type %v", caType)
	case UICAPem:
		return errors.Newf("unknown CA type %v", caType)
	default:
		return errors.Newf("unknown CA type %v", caType)
	}

	return nil
}

// CreateNodePair creates a node key and certificate.
// The CA cert and key must load properly. If multiple certificates
// exist in the CA cert, the first one is used.
func CreateNodePair(certsDir, caKeyPath string, keySize int, lifetime time.Duration, overwrite bool, hosts []string) error {
	if len(caKeyPath) == 0 {
		return errors.New("the path to the CA key is required")
	}
	if len(certsDir) == 0 {
		return errors.New("the path to the certs directory is required")
	}

	certsDirParam := fmt.Sprintf(CERTS_DIR, certsDir)
	caKeyParam := fmt.Sprintf(CA_KEY, caKeyPath)
	args := append(hosts, certsDirParam, caKeyParam)
	args = append([]string{CREATE_NODE}, args...)

	// run the crdb binary to generate the node certificates
	execCmd(args...)

	return nil
}

// CreateClientPair creates a node key and certificate.
// The CA cert and key must load properly. If multiple certificates
// exist in the CA cert, the first one is used.
// If a client CA exists, this is used instead.
// If wantPKCS8Key is true, the private key in PKCS#8 encoding is written as well.
func CreateClientPair(
	certsDir, caKeyPath string,
	keySize int,
	lifetime time.Duration,
	overwrite bool,
	user SQLUsername,
	wantPKCS8Key bool,
) error {
	if len(caKeyPath) == 0 {
		return errors.New("the path to the CA key is required")
	}
	if len(certsDir) == 0 {
		return errors.New("the path to the certs directory is required")
	}

	certsDirParam := fmt.Sprintf(CERTS_DIR, certsDir)
	caKeyParam := fmt.Sprintf(CA_KEY, caKeyPath)

	// TODO pks options do we need them?
	// run the crdb binary to generate the node certificates
	execCmd("create-client", user.U, certsDirParam, caKeyParam)

	return nil
}

// TODO should we run this??
// We do need the binary in our path

func LookPathCrdb() error {
	// we require the cockroach binary in the path
	_, err := exec.LookPath("cockroach")
	return err
}

// execCmd is a simple wrapper our exec that allows us to run a command
func execCmd(args ...string) {
	args = append([]string{CERT}, args...)
	cmd := exec.Command(CR, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		// TODO should we panic here or throw an error?
		// a panic will restart the pod
		panic(fmt.Sprintf("error: %s: %s\nout: %s\n", args, err, out))
	}
}
