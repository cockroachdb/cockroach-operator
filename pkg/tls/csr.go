package tls

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/cockroachlabs/crdb-operator/pkg/kube"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/pkg/errors"
	certs "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var Log = logf.Log.WithName("tls")

type RequestStatus string

const (
	SigningRequestPending  RequestStatus = "Pending"
	SigningRequestNotFound RequestStatus = "NotFound"
	SigningRequestApproved RequestStatus = "Approved"
	SigningRequestDenied   RequestStatus = "Denied"
)

func InitCSR(ctx context.Context, client client.Client, name string) (*CSR, error) {
	csr := &CSR{
		csr: &certs.CertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	err := kube.Get(ctx, client, csr.Unwrap())

	if kube.IgnoreNotFound(err) != nil {
		return nil, err
	}

	if kube.IsNotFound(err) {
		csr.Status = SigningRequestNotFound
		return csr, nil
	}

	for _, c := range csr.csr.Status.Conditions {
		if c.Type == certs.CertificateDenied {
			csr.Status = SigningRequestDenied
			return csr, nil
		}
		if c.Type == certs.CertificateApproved {
			csr.Status = SigningRequestApproved
			return csr, nil
		}
	}

	csr.Status = SigningRequestPending

	return csr, nil
}

type CSR struct {
	csr *certs.CertificateSigningRequest

	Status RequestStatus
}

func (c CSR) Unwrap() *certs.CertificateSigningRequest {
	return c.csr
}

func (c CSR) UnwrappedCopy() *certs.CertificateSigningRequest {
	return c.Unwrap().DeepCopy()
}

func NewNodeCertificateRequest(public, discovery, domain, namespace string) *x509.CertificateRequest {
	hosts := []string{
		"localhost",
		"127.0.0.1",
		public,
		fmt.Sprintf("%s.%s", public, namespace),
		fmt.Sprintf("%s.%s.%s", public, namespace, domain),
		fmt.Sprintf("*.%s", discovery),
		fmt.Sprintf("*.%s.%s", discovery, namespace),
		fmt.Sprintf("*.%s.%s.%s", discovery, namespace, domain),
	}

	req := &x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"Cockroach Labs"},
			CommonName:   "node",
		},
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			req.IPAddresses = append(req.IPAddresses, ip)
		} else {
			req.DNSNames = append(req.DNSNames, h)
		}
	}

	return req
}

func NewClientCertificateRequest(user string) *x509.CertificateRequest {
	return &x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"Cockroach Labs"},
			CommonName:   user,
		},
	}
}

func Approve(config *rest.Config, csr *certs.CertificateSigningRequest) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "failed to create kubernetes clientset")
	}

	updated := csr.DeepCopy()

	if updated.Status.Conditions == nil {
		updated.Status.Conditions = []certs.CertificateSigningRequestCondition{}
	}

	updated.Status.Conditions = append(updated.Status.Conditions,
		certs.CertificateSigningRequestCondition{
			Type:           certs.CertificateApproved,
			Reason:         "CRDBClusterTest",
			Message:        "Approved by CRDB operator",
			LastUpdateTime: metav1.Now(),
		},
	)

	// client-go can't properly handle this so clientset's help is required
	if _, err = clientset.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(updated); err != nil {
		return err
	}

	return nil
}

func SignAndCreate(request *x509.CertificateRequest,
	secret *resource.TLSSecret,
	csr *certs.CertificateSigningRequest,
	usages []certs.KeyUsage) error {
	pemKey, privateKey, err := initPrivateKey(secret)
	if err != nil {
		return err
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, request, privateKey)
	if err != nil {
		return err
	}

	req := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE REQUEST",
			Bytes: csrBytes,
		},
	)

	csr.Spec = certs.CertificateSigningRequestSpec{
		Request: req,
		Usages:  usages,
	}

	if err := secret.UpdateKey(pemKey); err != nil {
		return errors.Wrapf(err, "failed to update node TLS secret key")
	}

	return nil
}
