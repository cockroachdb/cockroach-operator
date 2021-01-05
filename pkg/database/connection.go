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

package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
)

const (
	CockroachDBSQLPort = 26257
	RootSQLUser        = "root"
)

// DBConnection represents a database connection into a CR Database
type DBConnection struct {
	Ctx context.Context
	// Client is the controller runtime client
	Client client.Client
	// RestConfig is the Kubernetes rest configuration
	RestConfig *rest.Config
	// ServiceName to connect to
	ServiceName string
	// Namespace that the pod is running in
	Namespace string
	// Database name that we connect to
	DatabaseName string
	// Port for the database connection
	Port *int32
	// RunningInsideK8s allows the database connection proxying
	// via a kube-proxy implementation
	RunningInsideK8s bool
	// UseSSL controls if the database connection utilizes SSL
	UseSSL bool
	// ClientCertificateSecretName is the name of the secret that contains the client CA and Key
	ClientCertificateSecretName string
	// RootCertificateSecretName is the name of the secret that contains the rootCA
	RootCertificateSecretName string
}

// NewDbConnection returns a new sql.DB instance to the corresponding CockroachDB pod.
// The DBConnection struct contains the information required to make the connection.
func NewDbConnection(dbConn *DBConnection) (*sql.DB, error) {

	c := &dbConfig{
		User: RootSQLUser,
		Host: dbConn.ServiceName,
		// ConnectTimeout was picked arbitrarily.
		// TODO have a way to set this
		ConnectTimeout: 15 * time.Second,
		Namespace:      dbConn.Namespace,
		Context:        dbConn.Ctx,
		Client:         dbConn.Client,
		Port:           int(*dbConn.Port),
		Database:       dbConn.DatabaseName,

		RunningInsideK8s: dbConn.RunningInsideK8s,
	}

	if dbConn.UseSSL {
		clientBundle, err := c.getClientTLSConfig(dbConn.ClientCertificateSecretName, dbConn.RootCertificateSecretName)
		if err != nil {
			return nil, errors.Wrap(err, "getting TLS certificate failed")
		}

		clientBundle.ServerName = dbConn.ServiceName
		c.TLSConfig = clientBundle
	}

	// We are Not Running Inside of K8s so use the dialer
	if !dbConn.RunningInsideK8s {
		podDialer, err := kube.NewPodDialer(dbConn.RestConfig, dbConn.Namespace)
		if err != nil {
			return nil, errors.Wrap(err, "creating new pod dialer failed")
		}
		dialerFunc := podDialer.DialContext
		c.DialFunc = &dialerFunc
		c.LookupFunc = lookupFunc
	}

	db, err := c.openDB()
	if err != nil {
		return nil, fmt.Errorf("opening a DB connection failed %s", err)
	}
	return db, nil
}

type dbConfig struct {
	Host string
	User string
	// Port defaults to CockroachDBSQLPort
	Port int
	// Database is the database being connected to.
	// it defaults to "system". This is the only database guaranteed to exist in CRDB clusters.
	Database string
	// DialFunc defaults to net.Dialer.DialContext. Set to kube.PodDailer.DialContext
	// if connecting to a k8s pod.
	DialFunc *func(context.Context, string, string) (net.Conn, error)
	// LookupFunc defaults to a stub, which is suitable for use with
	// kube.Dialer
	LookupFunc func(context.Context, string) ([]string, error)
	// TLSConfig describes the TLS configuration for this database connection.
	// nil disabled TLS connections.
	// CA certs should be specified via RootCAs
	// client certs should be specified via Certificates
	TLSConfig *tls.Config
	// ConnectTimeout restricts the whole connection process.
	ConnectTimeout time.Duration
	// Namespace that we are connecting to
	Namespace string
	// Context for the process
	Context context.Context
	// K8s Client
	Client client.Client

	// We use a dialer if we are not running inside K8s
	RunningInsideK8s bool
}

func (c dbConfig) getClientTLSConfig(clientCertificateSecretName string, rootCertificateSecretName string) (tlsConfig *tls.Config, err error) {
	r := resource.NewKubeResource(c.Context, c.Client, c.Namespace, kube.DefaultPersister)

	tlsSecret, err := resource.LoadTLSSecret(clientCertificateSecretName, r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load client tls secret")
	}

	keyPair, err := tls.X509KeyPair(tlsSecret.Key(), tlsSecret.PriveKey())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create key pair")
	}

	root, err := resource.LoadTLSSecret(rootCertificateSecretName, r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load root tls secret")
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(root.CA()) {
		return nil, fmt.Errorf("count not load CACert")
	}

	// Construct a tls.config
	return &tls.Config{
		Certificates: []tls.Certificate{keyPair},
		RootCAs:      pool,
	}, nil
}

func (c dbConfig) sslMode() string {
	if c.TLSConfig == nil {
		return "disable"
	}

	// TODO this is not used
	if c.TLSConfig.InsecureSkipVerify {
		return "verify-ca"
	}

	return "verify-full"
}

// openDB connects to a CockroachDB instance and verifies the connection
func (c dbConfig) openDB() (*sql.DB, error) {
	if c.Port == 0 {
		c.Port = CockroachDBSQLPort
	}

	if c.Database == "" {
		c.Database = "system"
	}

	connOptions := url.Values{
		"sslmode": {c.sslMode()},
	}

	pgURL := url.URL{
		Scheme:   "postgresql",
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     fmt.Sprintf("/%s", c.Database),
		RawQuery: connOptions.Encode(),
	}

	// pgx.Config's must be generated by ParseConfig
	// otherwise connect will panic.
	if c.User != "" {
		pgURL.User = url.User(c.User)
	}

	pgCfg, err := pgx.ParseConfig(pgURL.String())
	if err != nil {
		return nil, err
	}

	if !c.RunningInsideK8s {
		if c.DialFunc == nil {
			return nil, errors.New("dial func is not set")
		}
		pgCfg.DialFunc = *c.DialFunc
		pgCfg.LookupFunc = c.LookupFunc
	}

	pgCfg.TLSConfig = c.TLSConfig
	pgCfg.ConnectTimeout = c.ConnectTimeout

	db := stdlib.OpenDB(*pgCfg)

	// Test the database connection
	if _, err := db.Exec("SELECT 1"); err != nil {
		return nil, errors.Wrap(err, "testing db connection failed")
	}

	return db, nil
}

// lookupFunc is a stub for net.Resolver.Lookup. It's useful when connecting
// to a non-resolvable hostname, such as a kubernetes pod name.
func lookupFunc(ctx context.Context, host string) ([]string, error) {
	return []string{host}, nil
}
