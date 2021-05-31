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

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v2"
)

type target struct {
	template, output string
}

var targets = []target{
	{"manifests/operator.yaml.in", "manifests/operator.yaml"},
	{"manifests/patches/deployment_patch.yaml.in", "manifests/patches/deployment_patch.yaml"},
	{"config/samples/crdb-tls-example.yaml.in", "config/samples/crdb-tls-example.yaml"},
	{"examples/example.yaml.in", "examples/example.yaml"},
}

// crdb-versions.yaml structure
type crdbVersions struct {
	CrdbVersions []string `yaml:"CrdbVersions"`
}

type templateData struct {
	CrdbVersions      []Version
	LatestCrdbVersion string
	OperatorVersion   string
	GeneratedWarning  string
}

type Version struct {
	Version, UnderscoreVersion string
}

func main() {
	log.SetFlags(0)
	crdbVersionsFile := flag.String("crdb-versions", "", "YAML file with CRDB versions")
	operatorVersion := flag.String("operator-version", "", "Current operator version")
	repoRoot := flag.String("repo-root", "", "Git repository root")
	flag.Parse()

	if *crdbVersionsFile == "" || *operatorVersion == "" || *repoRoot == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	contents, err := ioutil.ReadFile(*crdbVersionsFile)
	if err != nil {
		log.Fatalf("Cannot open CRDB version file: %s", err)
	}
	var versions crdbVersions
	if err := yaml.Unmarshal(contents, &versions); err != nil {
		log.Fatalf("Cannot parse CRDB version file: %s", err)
	}
	// Collect all version in order to sort them
	vs := make([]*semver.Version, len(versions.CrdbVersions))
	for i, r := range versions.CrdbVersions {
		v, err := semver.NewVersion(r)
		if err != nil {
			log.Fatalf("Cannot convert version `%s`: %s", r, err)
		}
		vs[i] = v
	}
	sort.Sort(semver.Collection(vs))
	latestVersion := vs[len(vs)-1]

	var templateData templateData
	templateData.LatestCrdbVersion = latestVersion.Original()
	templateData.OperatorVersion = *operatorVersion
	templateData.GeneratedWarning = "Generated, do not edit"
	for _, v := range vs {
		templateData.CrdbVersions = append(templateData.CrdbVersions,
			Version{
				Version:           v.Original(),
				UnderscoreVersion: strings.ReplaceAll(v.Original(), ".", "_"),
			},
		)
	}

	for _, genData := range targets {
		if err := generateFile(templateData, *repoRoot, genData.template, genData.output); err != nil {
			log.Fatalf("Cannot generate %s: %s", genData.output, err)
		}
	}
}

func generateFile(templateData templateData, repoRoot, tplFile, output string) error {
	tpl := template.Must(template.ParseFiles(filepath.Join(repoRoot, tplFile)))
	outputAbs := filepath.Join(repoRoot, output)
	out, err := os.Create(outputAbs)
	if err != nil {
		return err
	}
	defer out.Close()
	return tpl.Execute(out, templateData)
}
