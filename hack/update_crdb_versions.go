package main

import (
	"bytes"
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
	"gopkg.in/yaml.v3"
)

const crdbVersionsInvertedRegexp = "^v19|^v21.1.8$|latest|ubi$"
const crdbVersionsFileName = "crdb-versions.yaml"

// TODO(rail): we may need to add pagination handling in case we pass 500 versions
// Use anonymous API to get the list of published images from the RedHat Catalog.
const crdbVersionsUrl = "https://catalog.redhat.com/api/containers/v1/repositories/registry/" +
	"registry.connect.redhat.com/repository/cockroachdb/cockroach/images?" +
	"exclude=data.repositories.comparison.advisory_rpm_mapping,data.brew," +
	"data.cpe_ids,data.top_layer_id&page_size=500&page=0"
const DefaultTimeout = 30
const crdbVersionsFileDescription = `#
# Supported CockroachDB versions.
#
# This file contains a list of CockroachDB versions that are supported by the
# operator. hack/crdbversions/main.go uses this list to generate various
# manifests.
# Please update this file when CockroachDB releases new versions.

`

type crdbVersionsResponse struct {
	Data []struct {
		Repositories []struct {
			Tags []struct {
				Name string `json:"name"`
			} `json:"tags"`
		} `json:"repositories"`
	} `json:"data"`
}

func main() {
	f, err := os.Create(crdbVersionsFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	responseData := crdbVersionsResponse{}
	err = getData(&responseData)
	if err != nil {
		log.Fatal(err)
	}

	// Get filtered and sorted versions in yaml representation
	versions := getVersions(responseData)
	sortedVersions := sortVersions(versions)
	yamlVersions := map[string][]string{"CrdbVersions": sortedVersions}

	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&yamlVersions)

	annotation := append(getCopyrightText(), []byte(crdbVersionsFileDescription)...)
	result := append(annotation, b.Bytes()...)
	err = ioutil.WriteFile(crdbVersionsFileName, result, 0)
	if err != nil {
		log.Fatal(err)
	}
}

func getData(data *crdbVersionsResponse) error {
	client := http.Client{Timeout: DefaultTimeout * time.Second}
	r, err := client.Get(crdbVersionsUrl)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(data)
}

func getVersions(data crdbVersionsResponse) []string {
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
	match, _ := regexp.MatchString(crdbVersionsInvertedRegexp, version)
	return !match
}

func sortVersions(versions []string) []string {
	vs := make([]*semver.Version, len(versions))
	for i, r := range versions {
		v, err := semver.NewVersion(r)
		if err != nil {
			fmt.Errorf("Error parsing version: %s", err)
		}

		vs[i] = v
	}
	sort.Sort(semver.Collection(vs))

	var sortedVersions []string
	for _, v := range vs {
		sortedVersions = append(sortedVersions, "v"+v.String())
	}
	return sortedVersions
}

func getCopyrightText() []byte {
	file, err := os.Open("hack/boilerplate/boilerplate.yaml.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
