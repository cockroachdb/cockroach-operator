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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v2"
)

const (
	httpTimeoutSecs = 30
	repo            = "registry.connect.redhat.com/cockroachdb/cockroach"
	versionsFile    = "crdb-versions.yaml"

	fileHeader = `#
# Supported CockroachDB versions.
#
# This file contains a list of CockroachDB versions that are supported by the
# operator. hack/crdbversions/main.go uses this list to generate various
# manifests.
# Please update this file when CockroachDB releases new versions.
#
# Generated. DO NOT EDIT. This file is created via make release/gen-templates

`

	// TODO(rail): we may need to add pagination handling in case we pass 500 versions
	// Use anonymous API to get the list of published images from the RedHat Catalog.
	reqPath = "/api/containers/v1/repositories/registry/registry.connect.redhat.com/" +
		"repository/cockroachdb/cockroach/images?" +
		"include=data.docker_image_digest,data.repositories&page_size=500&page=0"
)

var (
	// invalidVersions are known bad versions or non-semver versions that should not
	// be included in the results.
	invalidVersions = regexp.MustCompile("^v19|^v21.1.8$|latest|ubi$")

	// semVerRegex defines a Regexp for ensuring a valid (non-prerelease) version.
	semVerRegex = regexp.MustCompile(`v?([0-9]+)(\.[0-9]+)?(\.[0-9]+)?$`)
)

func main() {
	path := filepath.Join(os.Getenv("BUILD_WORKSPACE_DIRECTORY"), versionsFile)
	if err := os.WriteFile(path, []byte(fileHeader), 0644); err != nil {
		panic(err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	if err := UpdateCrdbVersions("https://catalog.redhat.com/", f); err != nil {
		panic(err)
	}
}

// UpdateCrdbVersions fetches all published versions of CRDB from the redhat
// connect registry. It then writes the results to the given io.Writer.
func UpdateCrdbVersions(baseURL string, w io.Writer) error {
	resp, err := fetchAPIResponse(fmt.Sprintf("%s%s", baseURL, reqPath))
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(generateOutput(resp))
	if err != nil {
		return errors.Wrap(err, "marshaling YAML")
	}

	if _, err := w.Write(data); err != nil {
		return errors.Wrap(err, "writing YAML")
	}

	return nil
}

func fetchAPIResponse(url string) (*apiResponse, error) {
	client := http.Client{Timeout: httpTimeoutSecs * time.Second}

	r, err := client.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "fetching CRDB versions")
	}
	defer r.Body.Close()

	var resp apiResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, errors.Wrap(err, "decoding API response")
	}

	return &resp, nil
}

func generateOutput(resp *apiResponse) *yamlOutput {
	output := new(yamlOutput)
	usedTags := make(map[string]bool)
	for _, data := range resp.Data {
		for _, r := range data.Repos {
			for _, tag := range r.Tags {
				if !isValid(tag.Name) || isUsed(usedTags, tag.Name) {
					continue
				}

				output.CrdbVersions = append(output.CrdbVersions, version{
					Image:       fmt.Sprintf("cockroachdb/cockroach:%s", tag.Name),
					RedhatImage: fmt.Sprintf("%s@%s", repo, data.Digest),
					Tag:         tag.Name,
				})
			}
		}
	}

	// ensure results are sorted properly by version
	sort.Slice(output.CrdbVersions, func(i, j int) bool {
		// safe to ignore error due to previous regexp check in isValid
		v1, _ := semver.NewVersion(output.CrdbVersions[i].Tag)
		v2, _ := semver.NewVersion(output.CrdbVersions[j].Tag)
		return v1.LessThan(v2)
	})

	return output
}

func isValid(tag string) bool {
	if invalidVersions.MatchString(tag) {
		return false
	}

	return semVerRegex.MatchString(tag)
}

// isUsed returns true if a tag has already been used to generate cockroach image.
func isUsed(usedTags map[string]bool, tag string) bool {
	if _, ok := usedTags[tag]; ok {
		return true
	}
	usedTags[tag] = true
	return false
}

// apiResponse encapsulates the response from the RH Catalog API.
type apiResponse struct {
	Data []struct {
		// Digest is used for digest pinning in OLM bundles (e.g. sha256@<sha>).
		Digest string `json:"docker_image_digest"`
		Repos  []struct {
			Tags []struct {
				// The image tag including the `v` prefix
				Name string `json:"name"`
			}
		} `json:"repositories"`
	} `json:"data"`
}

type yamlOutput struct {
	CrdbVersions []version `yaml:"CrdbVersions"`
}

type version struct {
	Image       string `yaml:"image"`
	RedhatImage string `yaml:"redhatImage"`
	Tag         string `yaml:"tag"`
}
