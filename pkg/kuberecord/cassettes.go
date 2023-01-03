/*
Copyright 2023 The Cockroach Authors

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

package kuberecord

import (
	"net/http"
	"net/url"

	"github.com/dnaeon/go-vcr/cassette"
)

const (
	// fingerPrintHeader is the header that checksums of outgoing requests are injected into. In replay mode, the
	// appropriate responses are found by matching on this header.
	fingerPrintHeader = "X-Kuberecord-Fingerprint"

	// kubernetesHost is a replacement for the kubernetes' API hostname. The hostname may change across test runs but that
	// doesn't invalidate our interactions. The recording library we use includes the HTTP Host header as a part of its
	// matching logic. Because of that, we need to have stable hostnames in our recordings. The easiest way to do that is
	// to simply rewrite the header. We don't care about the real name of whatever Kubernetes API server we're talking to.
	kubernetesHost = "KUBERNETES"
	schemeHTTPS    = "HTTPS"
)

var (
	prunableRequestHeaders  = []string{"Authorization"}
	prunableResponseHeaders = []string{"Date", "Audit-Id"}
)

// Filter modifies the Interaction before saving it to disk. This is necessary to normalize URLs and ensure that flakey
// and/or sensitive information isn't stored.
func Filter(i *cassette.Interaction) error {
	u, err := url.Parse(i.URL)
	if err != nil {
		return err
	}

	// Hostnames and schemes change across kubernetes clusters. Force all URLS to look like HTTPS://KUBERNETES/... to
	// make matching stable across different kubernetes clusters.
	u.Host = kubernetesHost
	u.Scheme = schemeHTTPS
	i.URL = u.String()

	// Sanitize outgoing requests. We don't want to accidentally check in credentials.
	for _, header := range prunableRequestHeaders {
		i.Request.Headers.Del(header)
	}

	// Remove any non-deterministic headers to reduce the churn of our records.
	for _, header := range prunableResponseHeaders {
		i.Response.Headers.Del(header)
	}

	return nil
}

// Matcher returns true when the incoming request matches a recorded interactions method, URL, and custom fingerprint
// header.
func Matcher(r *http.Request, i cassette.Request) bool {
	// Force all URLS to looks like https://KUBERNETES/... to make matching stable across different kubernetes clusters.
	// Hostnames change across kubernetes clusters and the empty rest.Config handed back in replay mode won't use HTTPS
	// unless it is provided with a TLS cert. Overridding the scheme here happens to be easier and just as affective.
	r.URL.Host = kubernetesHost
	r.URL.Scheme = schemeHTTPS

	// Method, URL, and custom fingerprint header must match.
	return cassette.DefaultMatcher(r, i) && r.Header.Get(fingerPrintHeader) == i.Headers.Get(fingerPrintHeader)
}
