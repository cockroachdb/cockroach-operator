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
	"fmt"
	"io"
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

// List of templates and destinations
var targets = []struct{ template, output string }{
	{"config/templates/operator.yaml.in", "manifests/operator.yaml"},
	{"config/templates/deployment_patch.yaml.in", "manifests/patches/deployment_patch.yaml"},
	{"config/templates/crdb-tls-example.yaml.in", "config/samples/crdb-tls-example.yaml"},
	{"config/templates/example.yaml.in", "examples/example.yaml"},
}

// crdb-versions.yaml structure
type crdbVersions struct {
	CrdbVersions []string `yaml:"CrdbVersions"`
}

type templateData struct {
	CrdbVersions            []*semver.Version
	LatestStableCrdbVersion string
	OperatorVersion         string
	GeneratedWarning        string
}

// readCrdbVersions reads CRDB versions from a YAML file and sorts them
// according to the semantic version sorting rules
func readCrdbVersions(r io.Reader) ([]*semver.Version, error) {
	contents, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot open CRDB version file: %w", err)
	}
	var versions crdbVersions
	if err := yaml.Unmarshal(contents, &versions); err != nil {
		return nil, fmt.Errorf("cannot parse CRDB version file: %w", err)
	}
	var vs []*semver.Version
	for _, raw := range versions.CrdbVersions {
		v, err := semver.NewVersion(raw)
		if err != nil {
			return nil, fmt.Errorf("cannot convert version `%s`: %w", r, err)
		}
		vs = append(vs, v)
	}
	sort.Sort(semver.Collection(vs))
	return vs, nil
}

func generateTemplateData(crdbVersions []*semver.Version, operatorVersion string) (templateData, error) {
	var data templateData
	data.OperatorVersion = operatorVersion
	data.GeneratedWarning = "Generated, do not edit"
	data.CrdbVersions = crdbVersions
	var stableVersions []*semver.Version
	for _, v := range data.CrdbVersions {
		if isStable(*v) {
			stableVersions = append(stableVersions, v)
		}
	}
	if len(stableVersions) == 0 {
		return templateData{}, fmt.Errorf("cannot find stable versions")
	}
	latestStable := stableVersions[len(stableVersions)-1]
	data.LatestStableCrdbVersion = latestStable.Original()
	return data, nil
}

func isStable(v semver.Version) bool {
	return v.Prerelease() == "" && v.Metadata() == ""
}

func dotsToUnderscore(v semver.Version) string {
	return strings.ReplaceAll(v.Original(), ".", "_")
}

func generateFile(name string, tplText string, output io.Writer, data templateData) error {
	// Teamplate functions
	funcs := template.FuncMap{
		"underscore": dotsToUnderscore,
		"stable":     isStable,
	}
	tpl, err := template.New(name).Funcs(funcs).Parse(tplText)
	if err != nil {
		return fmt.Errorf("cannot parse `%s`: %w", name, err)
	}
	return tpl.Execute(output, data)
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

	f, err := os.Open(*crdbVersionsFile)
	if err != nil {
		log.Fatalf("Cannot open versions file: %s", err)
	}
	defer f.Close()

	vs, err := readCrdbVersions(f)
	if err != nil {
		log.Fatalf("Cannot read versions file: %s", err)
	}

	data, err := generateTemplateData(vs, *operatorVersion)
	if err != nil {
		log.Fatalf("Cannot generate template data: %s", err)
	}

	for _, f := range targets {
		tplFile := filepath.Join(*repoRoot, f.template)
		outputFile := filepath.Join(*repoRoot, f.output)
		name := filepath.Base(outputFile)
		tplContents, err := ioutil.ReadFile(tplFile)
		if err != nil {
			log.Fatalf("Cannot read template file `%s`: %s", tplFile, err)
		}
		output, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Cannot create `%s`: %s", outputFile, err)
		}
		defer output.Close()

		if err := generateFile(name, string(tplContents), output, data); err != nil {
			log.Fatalf("Cannot generate %s: %s", f.output, err)
		}
	}
}
