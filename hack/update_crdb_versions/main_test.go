/*
Copyright 2022 The Cockroach Authors

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

package main_test

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/cockroachdb/cockroach-operator/hack/update_crdb_versions"
	"github.com/stretchr/testify/require"
)

func TestUpdateCrdbVersions(t *testing.T) {
	images := []struct {
		Note string
		Sha  string
		Tag  string
	}{
		// These are in expected order
		{Sha: "sha256:image1", Tag: "v1"},
		{Sha: "sha256:image1.2", Tag: "v1.2"},
		{Sha: "sha256:image1.10", Tag: "v1.10"},
		{Sha: "sha256:image2", Tag: "v2"},
		{Note: "v19* not supported", Tag: "v19.0.1"},
		{Note: "v21.1.8 has an issue with rollbacks", Tag: "v21.1.8"},
		{Note: "latest isn't stable", Tag: "latest"},
		{Note: "ubi is not wanted", Tag: "ubi"},
		{Note: "prerelease not suppored", Tag: "v1-alpha"},
		{Note: "metadata not supported", Tag: "v1+snapshot"},
	}

	tmpl := template.Must(template.New("rhAPI").Parse(`
{
  "data": [
{{ range $index, $el:= . }}
  {{ if $index }},{{ end }}
  {
    "docker_image_digest": "{{ $el.Sha }}",
    "repositories": [
      { "tags": [{ "name": "{{ $el.Tag }}" }] }
    ]
  }
{{ end }}
  ]
}
`))

	var expected strings.Builder
	expected.WriteString("CrdbVersions:\n")
	for _, img := range images {
		if img.Sha != "" {
			expected.WriteString(fmt.Sprintf("- image: cockroachdb/cockroach:%s\n", img.Tag))
			expected.WriteString(fmt.Sprintf("  redhatImage: registry.connect.redhat.com/cockroachdb/cockroach@%s\n", img.Sha))
			expected.WriteString(fmt.Sprintf("  tag: %s\n", img.Tag))
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		// shuffle images to ensure semver sort is working
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })

		tmpl.Execute(w, images)
	}))
	defer server.Close()

	var str strings.Builder
	require.NoError(t, UpdateCrdbVersions(server.URL, &str))
	require.Equal(t, expected.String(), str.String())
}
