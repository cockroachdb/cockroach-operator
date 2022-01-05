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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v2"
)

const crdbVersionsInvertedRegexp = "^v19|^v21.1.8$|latest|ubi$"
const CrdbVersionsFileName = "crdb-versions.yaml"

// TODO(rail): we may need to add pagination handling in case we pass 500 versions
// Use anonymous API to get the list of published images from the RedHat Catalog.
const crdbVersionsUrl = "https://catalog.redhat.com/api/containers/v1/repositories/registry/" +
	"registry.connect.redhat.com/repository/cockroachdb/cockroach/images?" +
	"exclude=data.repositories.comparison.advisory_rpm_mapping,data.brew," +
	"data.cpe_ids,data.top_layer_id&page_size=500&page=0"
const crdbVersionsDefaultTimeout = 30
const CrdbVersionsFileDescription = `#
# Supported CockroachDB versions.
#
# This file contains a list of CockroachDB versions that are supported by the
# operator. hack/crdbversions/main.go uses this list to generate various
# manifests.
# Please update this file when CockroachDB releases new versions.
#
# Generated. DO NOT EDIT. This file is created via make release/gen-templates

`

type CrdbVersionsResponse struct {
	Data []struct {
		Repositories []struct {
			Tags []struct {
				Name string `json:"name"`
			} `json:"tags"`
		} `json:"repositories"`
	} `json:"data"`
}

func GetData(data *CrdbVersionsResponse) error {
	client := http.Client{Timeout: crdbVersionsDefaultTimeout * time.Second}
	r, err := client.Get(crdbVersionsUrl)
	if err != nil {
		return fmt.Errorf("Cannot make a get request: %s", err)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(data)
}

func GetVersions(data CrdbVersionsResponse) []string {
	var versions []string
	for _, data := range data.Data {
		for _, repo := range data.Repositories {
			for _, tag := range repo.Tags {
				if IsValid(tag.Name) {
					versions = append(versions, tag.Name)
				}
			}
		}
	}
	return versions
}

func IsValid(version string) bool {
	match, _ := regexp.MatchString(crdbVersionsInvertedRegexp, version)
	return !match
}

// sortVersions converts the slice with versions to slice with semver.Version
// sorts them and converts back to slice with version strings
func SortVersions(versions []string) []string {
	vs := make([]*semver.Version, len(versions))
	for i, r := range versions {
		v, err := semver.NewVersion(r)
		if err != nil {
			log.Fatalf("Cannot parse version : %s", err)
		}

		vs[i] = v
	}
	sort.Sort(semver.Collection(vs))

	var sortedVersions []string
	for _, v := range vs {
		sortedVersions = append(sortedVersions, v.Original())
	}
	return sortedVersions
}

func GenerateCrdbVersionsFile(versions []string, path string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Cannot create %s file: %s", path, err)
	}
	defer f.Close()

	versionsMap := map[string][]string{"CrdbVersions": versions}
	yamlVersions, err := yaml.Marshal(&versionsMap)
	if err != nil {
		log.Fatalf("error while converting to yaml: %v", err)
	}

	result := append([]byte(CrdbVersionsFileDescription), yamlVersions...)
	return ioutil.WriteFile(path, result, 0)
}

func main() {
	responseData := CrdbVersionsResponse{}
	err := GetData(&responseData)
	if err != nil {
		log.Fatalf("Cannot parse response: %s", err)
	}

	rawVersions := GetVersions(responseData)
	sortedVersions := SortVersions(rawVersions)

	err = GenerateCrdbVersionsFile(sortedVersions, CrdbVersionsFileName)
	if err != nil {
		log.Fatalf("Cannot write %s file: %s", CrdbVersionsFileName, err)
	}
}
