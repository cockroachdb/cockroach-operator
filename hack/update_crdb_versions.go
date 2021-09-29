package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const INVERTED_VERSIONS_REGEXP = "^v19|^v21.1.8$|latest|ubi$"
const FILE_NAME = "crdb-versions.yaml"

// TODO(rail): we may need to add pagination handling in case we pass 500 versions
// Use anonymous API to get the list of published images from the RedHat Catalog.
const URL = "https://catalog.redhat.com/api/containers/v1/repositories/registry/" +
	"registry.connect.redhat.com/repository/cockroachdb/cockroach/images?" +
	"exclude=data.repositories.comparison.advisory_rpm_mapping,data.brew," +
	"data.cpe_ids,data.top_layer_id&page_size=500&page=0"

const COPYRIGHT = `# Copyright 2021 The Cockroach Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Supported CockroachDB versions.
#
# This file contains a list of CockroachDB versions that are supported by the
# operator. hack/crdbversions/main.go uses this list to generate various
# manifests.
# Please update this file when CockroachDB releases new versions.

`

type Response struct {
	Data []struct {
		Repositories []struct {
			Tags []struct {
				Name string `json:"name"`
			} `json:"tags"`
		} `json:"repositories"`
	} `json:"data"`
}

func main() {
	_, err := os.Create(FILE_NAME)
	if err != nil {
		log.Fatal(err)
	}

	responseData := Response{}
	err = getData(&responseData)
	if err != nil {
		log.Fatal(err)
	}

	// Get filtered and sorted versions in yaml representation
	versions := getVersions(responseData)
	sortedVersions := sort(versions)
	yamlVersions := map[string][]string{"CrdbVersions": sortedVersions}

	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&yamlVersions)

	result := append([]byte(COPYRIGHT), b.Bytes()...)
	err = ioutil.WriteFile(FILE_NAME, result, 0)
	if err != nil {
		log.Fatal(err)
	}
}

func getData(data *Response) error {
	r, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(data)
}

func getVersions(data Response) []string {
	var versions []string
	for _, data := range data.Data {
		for _, repo := range data.Repositories {
			for _, tag := range repo.Tags {
				if validVersion(tag.Name) {
					versions = append(versions, tag.Name)
				}
			}
		}
	}
	return versions
}

func validVersion(version string) bool {
	match, _ := regexp.MatchString(INVERTED_VERSIONS_REGEXP, version)
	return !match
}

func sort(versions []string) []string {
	cmd := "echo " + "\"\n" + strings.Join(versions, "\n") + "\"" + " | sort -V"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Sprintf("Failed to execute command: %s", cmd)
	}
	splitFn := func(c rune) bool {
		return c == '\n'
	}
	return strings.FieldsFunc(string(out), splitFn)
}
