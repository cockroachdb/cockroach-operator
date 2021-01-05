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

package kube

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// podAddr implements net.Addr for kubernetes pod connections
type podAddr struct {
	PodName string
	Port    string
}

func (podAddr) Network() string {
	return "tcp"
}

func (a podAddr) String() string {
	return fmt.Sprintf("%s:%s", a.PodName, a.Port)
}

// PodDialer uses kubernetes' portforwarding protocol to create a net.Conn
// to a pod in the given kubernetes clusters
type PodDialer struct {
	Namespace string
	Config    *rest.Config
	ClientSet kubernetes.Interface
	Transport http.RoundTripper
	Upgrader  spdy.Upgrader

	mu             sync.Mutex
	requestCounter int
	dialers        map[string]httpstream.Dialer
}

// NewPodDialer creates a PodDailer that allows for a database connection to flow
// through a connection created by a kube-proxy like connection.
func NewPodDialer(config *rest.Config, namespace string) (*PodDialer, error) {
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	return &PodDialer{
		Config:    config,
		Namespace: namespace,
		ClientSet: kubernetes.NewForConfigOrDie(config),
		Transport: transport,
		Upgrader:  upgrader,
		dialers:   make(map[string]httpstream.Dialer),
	}, nil
}

func (k *PodDialer) nextRequestID() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.requestCounter += 1
	return k.requestCounter
}

func (d *PodDialer) dialerForPod(podName string) httpstream.Dialer {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Reuse any dialers that are already created
	if dialer, ok := d.dialers[podName]; ok {
		return dialer
	}

	podShortName := strings.Split(podName, ".")

	// Build a raw request so we can extract the URL
	req := d.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(d.Namespace).
		Name(podShortName[0]).
		SubResource("portforward")

	dialer := spdy.NewDialer(d.Upgrader, &http.Client{Transport: d.Transport}, "POST", req.URL())

	d.dialers[podName] = dialer

	return dialer
}

// DialTimeout implements pq.Dialer
func (k *PodDialer) DialTimeout(network, addr string, d time.Duration) (net.Conn, error) {
	return k.Dial(network, addr)
}

func (k *PodDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return k.Dial(network, addr)
}

// Dial connects to a port in a kubernetes pod specified by addr. network must be TCP
// Implmentation adapted from:
//	https://github.com/kubernetes/kubernetes/blob/27c70773add99e43464a4e525e3bddfc8b602a3d/staging/src/k8s.io/client-go/tools/portforward/portforward.go
//	https://github.com/kubernetes/kubernetes/blob/27c70773add99e43464a4e525e3bddfc8b602a3d/staging/src/k8s.io/kubectl/pkg/cmd/portforward/portforward.go
func (k *PodDialer) Dial(network, addr string) (net.Conn, error) {
	if network != "tcp" {
		return nil, errors.New("only tcp networks are currently supported")
	}

	podName, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	dialer := k.dialerForPod(podName)

	streamConn, _, err := dialer.Dial(portforward.PortForwardProtocolV1Name)
	if err != nil {
		return nil, err
	}

	requestID := k.nextRequestID()

	headers := http.Header{}
	headers.Set(corev1.StreamType, corev1.StreamTypeError)
	headers.Set(corev1.PortHeader, port)
	headers.Set(corev1.PortForwardRequestIDHeader, fmt.Sprintf("%d", requestID))

	errStream, err := streamConn.CreateStream(headers)
	if err != nil {
		return nil, err
	}

	// we're not writing to this stream
	errStream.Close()

	errorChan := make(chan error)
	go func() {
		message, err := ioutil.ReadAll(errStream)
		if err != nil {
			errorChan <- err
		}
		if len(message) > 0 {
			errorChan <- fmt.Errorf("%s", message)
		}
		close(errorChan)
	}()

	// create data stream
	headers.Set(corev1.StreamType, corev1.StreamTypeData)
	dataStream, err := streamConn.CreateStream(headers)
	if err != nil {
		return nil, err
	}

	// dataStream is expected to be a *spdystream.Stream which implements net.Conn.
	// We're leaving open the option to use other transports as long as they implement net.Conn as well.
	if conn, ok := dataStream.(net.Conn); ok {
		return &podConn{
			Conn:      conn,
			PodName:   podName,
			Port:      port,
			errorChan: errorChan,
		}, nil
	}

	// Ignore error from Close() as we're just trying to clean up
	_ = dataStream.Close()

	return nil, errors.New("datastream does not implement net.Conn")
}

type podConn struct {
	net.Conn

	errorChan <-chan error

	PodName string
	Port    string
}

func (c *podConn) RemoteAddr() net.Addr {
	return podAddr{PodName: c.PodName, Port: c.Port}
}

func (c *podConn) Close() error {
	if err := c.Conn.Close(); err != nil {
		return err
	}

	// Ensure that this connection hasn't terminated abnormally
	return <-c.errorChan
}
