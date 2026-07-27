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

package main_test

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/cockroachdb/cockroach-operator/hack/update_crdb_versions"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	crdbURLPath = "/v2/cockroachdb/cockroach/tags"
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

	dockerImages := []struct {
		Note string
		Sha  string
		Tag  string
	}{
		// These are in expected (semver) order. v1.10 sorts after v1.2, so it
		// distinguishes semantic ordering from lexical ordering.
		{Sha: "sha256:image1", Tag: "v1"},
		{Sha: "sha256:image1.2", Tag: "v1.2"},
		{Sha: "sha256:image1.10", Tag: "v1.10"},
		{Sha: "sha256:image2", Tag: "v2"},
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
	for _, img := range dockerImages {
		if img.Sha != "" {
			fmt.Fprintf(&expected, "- image: cockroachdb/cockroach:%s\n", img.Tag)
			fmt.Fprintf(&expected, "  redhatImage: registry.connect.redhat.com/cockroachdb/cockroach@%s\n", img.Sha)
			fmt.Fprintf(&expected, "  tag: %s\n", img.Tag)
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		// shuffle images to ensure semver sort is working
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		rnd.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })

		require.NoError(t, tmpl.Execute(w, images))
	}))
	defer server.Close()

	var dockerServer *httptest.Server
	dockerRequests := 0
	dockerServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, crdbURLPath, r.URL.Path)
		dockerRequests++

		// Exercise retry handling without slowing down the test.
		if dockerRequests == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		results := dockerImages[2:]
		next := ""
		if r.URL.Query().Get("page") == "" {
			results = dockerImages[:2]
			next = dockerServer.URL + crdbURLPath + "?page=2"
		}

		names := make([]string, 0, len(results))
		for _, img := range results {
			names = append(names, img.Tag)
		}
		writeDockerPage(t, w, len(dockerImages), next, names)
	}))
	defer dockerServer.Close()

	previousDockerHubURL := BaseDockerHubURL
	BaseDockerHubURL = dockerServer.URL + crdbURLPath
	t.Cleanup(func() { BaseDockerHubURL = previousDockerHubURL })

	updated, err := GenerateCrdbVersions(server.URL)
	require.NoError(t, err)
	data, err := yaml.Marshal(updated)
	require.NoError(t, err)
	require.Equal(t, expected.String(), string(data))
	require.Equal(t, 3, dockerRequests, "one retry plus two paginated requests")
}

func TestUpdateCrdbVersionsFailsWhenDockerHubRemainsUnavailable(t *testing.T) {
	redHatServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer redHatServer.Close()

	dockerRequests := 0
	dockerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dockerRequests++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer dockerServer.Close()

	previousDockerHubURL := BaseDockerHubURL
	BaseDockerHubURL = dockerServer.URL + crdbURLPath
	t.Cleanup(func() { BaseDockerHubURL = previousDockerHubURL })

	_, err := GenerateCrdbVersions(redHatServer.URL)
	require.ErrorContains(t, err, "request failed after 6 attempts")
	require.Equal(t, 6, dockerRequests)
}

// TestFetchDockerHubTagsRetriesMidPagination ensures a 429 encountered after the
// first page has already been collected is retried without losing tags.
func TestFetchDockerHubTagsRetriesMidPagination(t *testing.T) {
	redHatServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[
			{"docker_image_digest":"sha256:image1","repositories":[{"tags":[{"name":"v1"}]}]},
			{"docker_image_digest":"sha256:image2","repositories":[{"tags":[{"name":"v2"}]}]}
		]}`)
	}))
	defer redHatServer.Close()

	var dockerServer *httptest.Server
	page2Attempts := 0
	dockerServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "" {
			writeDockerPage(t, w, 2, dockerServer.URL+crdbURLPath+"?page=2", tagNames("v1"))
			return
		}

		// The first time page two is requested, force a retry mid-pagination.
		page2Attempts++
		if page2Attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		writeDockerPage(t, w, 2, "", tagNames("v2"))
	}))
	defer dockerServer.Close()

	previousDockerHubURL := BaseDockerHubURL
	BaseDockerHubURL = dockerServer.URL + crdbURLPath
	t.Cleanup(func() { BaseDockerHubURL = previousDockerHubURL })

	updated, err := GenerateCrdbVersions(redHatServer.URL)
	require.NoError(t, err)
	require.Equal(t, 2, page2Attempts, "page two should be retried once")
	require.Equal(t, []string{"v1", "v2"}, versionTags(updated))
}

// TestFetchDockerHubTagsDetectsPaginationLoop ensures a self-referential Next
// pointer is detected instead of looping forever.
func TestFetchDockerHubTagsDetectsPaginationLoop(t *testing.T) {
	redHatServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer redHatServer.Close()

	var dockerServer *httptest.Server
	dockerServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always point Next at the same URL to create a loop. count is left at
		// zero so the completeness check does not fire first.
		writeDockerPage(t, w, 0, dockerServer.URL+crdbURLPath+"?page=stuck", tagNames("v1"))
	}))
	defer dockerServer.Close()

	previousDockerHubURL := BaseDockerHubURL
	BaseDockerHubURL = dockerServer.URL + crdbURLPath
	t.Cleanup(func() { BaseDockerHubURL = previousDockerHubURL })

	_, err := GenerateCrdbVersions(redHatServer.URL)
	require.ErrorContains(t, err, "Docker Hub pagination loop")
}

// TestUpdateVersionsFileKeepsFileOnRemoval proves the end-to-end flow leaves the
// versions file untouched when generation would drop a supported version, and
// that -allow-removals lets the removal through.
func TestUpdateVersionsFileKeepsFileOnRemoval(t *testing.T) {
	// Generation only sees v1 and v2, so v1.2 would be removed.
	redHatServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[
			{"docker_image_digest":"sha256:image1","repositories":[{"tags":[{"name":"v1"}]}]},
			{"docker_image_digest":"sha256:image2","repositories":[{"tags":[{"name":"v2"}]}]}
		]}`)
	}))
	defer redHatServer.Close()

	dockerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeDockerPage(t, w, 2, "", tagNames("v1", "v2"))
	}))
	defer dockerServer.Close()

	previousDockerHubURL := BaseDockerHubURL
	BaseDockerHubURL = dockerServer.URL + crdbURLPath
	t.Cleanup(func() { BaseDockerHubURL = previousDockerHubURL })

	original := `#
# Supported CockroachDB versions.
#
CrdbVersions:
- image: cockroachdb/cockroach:v1
  redhatImage: registry.connect.redhat.com/cockroachdb/cockroach@sha256:image1
  tag: v1
- image: cockroachdb/cockroach:v1.2
  redhatImage: registry.connect.redhat.com/cockroachdb/cockroach@sha256:image1.2
  tag: v1.2
- image: cockroachdb/cockroach:v2
  redhatImage: registry.connect.redhat.com/cockroachdb/cockroach@sha256:image2
  tag: v2
`
	path := filepath.Join(t.TempDir(), "crdb-versions.yaml")
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	// Without -allow-removals the run aborts and the file is left as-is.
	err := UpdateVersionsFile(path, redHatServer.URL, false)
	require.ErrorContains(t, err, "refusing to remove existing CRDB versions: v1.2")

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, original, string(after), "file must be unchanged when a removal is refused")

	// With -allow-removals the removal is permitted and the file is rewritten.
	require.NoError(t, UpdateVersionsFile(path, redHatServer.URL, true))
	rewritten, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NotContains(t, string(rewritten), "tag: v1.2")
	require.Contains(t, string(rewritten), "tag: v1\n")
	require.Contains(t, string(rewritten), "tag: v2\n")
}

// TestUpdateVersionsFileBootstrapsMissingFile ensures a missing versions file is
// treated as an empty baseline rather than a fatal error.
func TestUpdateVersionsFileBootstrapsMissingFile(t *testing.T) {
	redHatServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[
			{"docker_image_digest":"sha256:image1","repositories":[{"tags":[{"name":"v1"}]}]}
		]}`)
	}))
	defer redHatServer.Close()

	dockerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeDockerPage(t, w, 1, "", tagNames("v1"))
	}))
	defer dockerServer.Close()

	previousDockerHubURL := BaseDockerHubURL
	BaseDockerHubURL = dockerServer.URL + crdbURLPath
	t.Cleanup(func() { BaseDockerHubURL = previousDockerHubURL })

	path := filepath.Join(t.TempDir(), "crdb-versions.yaml")
	require.NoError(t, UpdateVersionsFile(path, redHatServer.URL, false))

	written, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(written), "tag: v1\n")
}

// TestFetchDockerHubTagsDetectsTruncatedListing ensures a listing that reports
// more tags than it returns is rejected instead of silently dropping versions.
func TestFetchDockerHubTagsDetectsTruncatedListing(t *testing.T) {
	redHatServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer redHatServer.Close()

	dockerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reports 5 total but returns only 1 tag and no further pages.
		writeDockerPage(t, w, 5, "", tagNames("v1"))
	}))
	defer dockerServer.Close()

	previousDockerHubURL := BaseDockerHubURL
	BaseDockerHubURL = dockerServer.URL + crdbURLPath
	t.Cleanup(func() { BaseDockerHubURL = previousDockerHubURL })

	_, err := GenerateCrdbVersions(redHatServer.URL)
	require.ErrorContains(t, err, "incomplete Docker Hub tag listing")
}

func TestValidateNoVersionRemovals(t *testing.T) {
	existing := []string{"v25.2.2", "v26.2.4"}

	require.NoError(t, ValidateNoVersionRemovals(existing,
		[]string{"v25.2.2", "v26.2.4", "v26.2.5"}))

	err := ValidateNoVersionRemovals(existing, []string{"v26.2.4"})
	require.EqualError(t, err,
		"refusing to remove existing CRDB versions: v25.2.2; rerun with CRDB_VERSION_UPDATE_ARGS=-allow-removals for an intentional removal")
}

// writeDockerPage encodes a single Docker Hub tags page. count is the total the
// API claims to hold; next is the URL of the following page ("" when last).
func writeDockerPage(t *testing.T, w http.ResponseWriter, count int, next string, names []string) {
	t.Helper()
	response := struct {
		Count   int    `json:"count"`
		Next    string `json:"next"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}{Count: count, Next: next}
	for _, name := range names {
		response.Results = append(response.Results, struct {
			Name string `json:"name"`
		}{Name: name})
	}
	require.NoError(t, json.NewEncoder(w).Encode(response))
}

func tagNames(names ...string) []string {
	return names
}

// versionTags marshals generated versions and extracts just the tags in order.
func versionTags(updated any) []string {
	data, err := yaml.Marshal(updated)
	if err != nil {
		panic(err)
	}
	var out struct {
		CrdbVersions []struct {
			Tag string `yaml:"tag"`
		} `yaml:"CrdbVersions"`
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		panic(err)
	}
	tags := make([]string, 0, len(out.CrdbVersions))
	for _, v := range out.CrdbVersions {
		tags = append(tags, v.Tag)
	}
	return tags
}
