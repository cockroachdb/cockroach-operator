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

package kuberecord_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/kuberecord"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	interaction := &cassette.Interaction{
		Request: cassette.Request{
			Method: "GET",
			URL:    "http://localhost:8080/foo?bar=baz",
			Headers: http.Header{
				// Authorization header will be removed
				"Authorization": []string{"Bearer bogus-token"},
				"Accept":        []string{"application/json"},
			},
		},
		Response: cassette.Response{
			Headers: http.Header{
				// Date and Audit-Id headers not recorded
				"Date":           []string{"2020-08-17T00.00.000Z"},
				"Audit-Id":       []string{"12345abcde"},
				"X-Custom-Value": []string{"Custom Value"},
			},
		},
	}

	require.Nil(t, Filter(interaction))
	require.Equal(t, "HTTPS://KUBERNETES/foo?bar=baz", interaction.Request.URL)
	require.Empty(t, interaction.Request.Headers.Get("Authorization"))
	require.Equal(t, "application/json", interaction.Request.Headers.Get("Accept"))

	require.Empty(t, interaction.Response.Headers.Get("Date"))
	require.Empty(t, interaction.Response.Headers.Get("Audit-Id"))
	require.Equal(t, "Custom Value", interaction.Response.Headers.Get("X-Custom-Value"))
}

func TestMatcher(t *testing.T) {
	fpHeader := "X-Kuberecord-Fingerprint"
	makeReq := func(method, url, fp string) *http.Request {
		req := httptest.NewRequest(method, url, nil)
		req.Header.Set(fpHeader, fp)
		return req
	}

	makeCassette := func(method, url, fp string) cassette.Request {
		return cassette.Request{
			Method:  method,
			URL:     url,
			Headers: http.Header{fpHeader: []string{fp}},
		}
	}

	tests := []struct {
		name     string
		req      *http.Request
		cassette cassette.Request
		pass     bool
	}{
		{
			name:     "perfect match",
			req:      makeReq(http.MethodPost, "HTTPS://KUBERNETES", "some-sha"),
			cassette: makeCassette(http.MethodPost, "HTTPS://KUBERNETES", "some-sha"),
			pass:     true,
		},
		{
			name:     "fuzzy match with different URLs",
			req:      makeReq(http.MethodPost, "https://localhost:8080/test", "some-sha"),
			cassette: makeCassette(http.MethodPost, "HTTPS://KUBERNETES/test", "some-sha"),
			pass:     true,
		},
		{
			name:     "missing fingerprint",
			req:      makeReq(http.MethodPost, "https://localhost:8080/test", ""),
			cassette: makeCassette(http.MethodPost, "HTTPS://KUBERNETES/test", "some-sha"),
		},
		{
			name:     "mismatched fingerprint",
			req:      makeReq(http.MethodPost, "https://localhost:8080/test", "some-other-sha"),
			cassette: makeCassette(http.MethodPost, "HTTPS://KUBERNETES/test", "some-sha"),
		},
		{
			name:     "mismatched URL",
			req:      makeReq(http.MethodPost, "https://localhost:8080/test", "some-sha"),
			cassette: makeCassette(http.MethodPost, "http://KUBERNETES/test", "some-sha"),
		},
	}

	for _, tt := range tests {
		require.Equal(t, tt.pass, Matcher(tt.req, tt.cassette), tt.name)
	}
}
