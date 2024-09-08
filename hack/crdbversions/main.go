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

// This program generates various YAML files based on predefined list of
// templates. crdb-versions.yaml is used to define supported CockroachDB
// versions.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

// List of templates and destinations
var targets = []struct{ template, output string }{
	{"config/templates/client-secure-operator.yaml.in", "examples/client-secure-operator.yaml"},
	{"config/templates/crdb-tls-example.yaml.in", "config/samples/crdb-tls-example.yaml"},
	{"config/templates/deployment_image.yaml.in", "config/manager/patches/image.yaml"},
	{"config/templates/deployment_patch.yaml.in", "config/manifests/patches/deployment_patch.yaml"},
	{"config/templates/example.yaml.in", "examples/example.yaml"},
	{"config/templates/smoketest.yaml.in", "examples/smoketest.yaml"},
	{"config/templates/csv.yaml.in", "config/manifests/bases/cockroach-operator.clusterserviceversion.yaml"},
}

// crdb-versions.yaml structure
type crdbVersions struct {
	CrdbVersions []crdbVersion `yaml:"CrdbVersions"`
}

type crdbVersion struct {
	Image       string `yaml:"image"`
	RedhatImage string `yaml:"redhatImage"`
	Tag         string `yaml:"tag"`
}

type templateData struct {
	CrdbVersions            []crdbVersion
	LatestStableCrdbVersion string
	OperatorImage           string
	OperatorVersion         string
	GeneratedWarning        string
	Year                    string
}

func main() {
	log.SetFlags(0)
	crdbVersionsFile := flag.String("crdb-versions", "", "YAML file with CRDB versions")
	operatorVersion := flag.String("operator-version", "", "Current operator version")
	operatorImage := flag.String("operator-image", "cockroachdb/cockroach-operator", "Current operator image")
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

	data, err := generateTemplateData(vs, *operatorVersion, *operatorImage)
	if err != nil {
		log.Fatalf("Cannot generate template data: %s", err)
	}

	for _, f := range targets {
		tplFile := filepath.Join(*repoRoot, f.template)
		outputFile := filepath.Join(*repoRoot, f.output)
		log.Printf("generating `%s` from `%s`", outputFile, tplFile)
		name := filepath.Base(outputFile)
		tplContents, err := os.ReadFile(tplFile)
		if err != nil {
			log.Fatalf("Cannot read template file `%s`: %s", tplFile, err)
		}
		if err := os.MkdirAll(filepath.Dir(outputFile), 0700); err != nil {
			log.Fatalf("Cannot create directory for %s", outputFile)
		}

		output, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Cannot create `%s`: %s", outputFile, err)
		}

		data.GeneratedWarning = fmt.Sprintf("Generated, do not edit. Please edit this file instead: %s", f.template)
		if err := generateFile(name, string(tplContents), output, data); err != nil {
			log.Fatalf("Cannot generate %s: %s", f.output, err)
		}
		if err := output.Close(); err != nil {
			log.Fatalf("Cannot close `%s`: %s", outputFile, err)
		}
		log.Printf("verifying `%s`", outputFile)
		if err := verifyYamlLoads(outputFile); err != nil {
			log.Fatalf("Cannot load YAML `%s`: %s", outputFile, err)
		}
	}
}

func readCrdbVersions(r io.Reader) ([]crdbVersion, error) {
	// nolint
	// io.ReadAll is available from 1.16 onwards. Earlier it was available in io/ioutil package
	contents, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cannot open CRDB version file: %w", err)
	}
	var versions crdbVersions
	if err := yaml.Unmarshal(contents, &versions); err != nil {
		return nil, fmt.Errorf("cannot parse CRDB version file: %w", err)
	}

	return versions.CrdbVersions, nil
}

func generateTemplateData(crdbVersions []crdbVersion, operatorVersion, operatorImage string) (templateData, error) {
	var data templateData
	data.Year = fmt.Sprint(time.Now().Year())
	data.OperatorVersion = operatorVersion
	data.OperatorImage = operatorImage
	data.CrdbVersions = crdbVersions

	latestStable := crdbVersions[len(crdbVersions)-1].Tag
	data.LatestStableCrdbVersion = latestStable
	return data, nil
}

func dotsToUnderscore(v string) string {
	return strings.ReplaceAll(v, ".", "_")
}

func generateFile(name string, tplText string, output io.Writer, data templateData) error {
	// Template functions
	funcs := template.FuncMap{
		"underscore": dotsToUnderscore,
	}
	tpl, err := template.New(name).Funcs(funcs).Parse(tplText)
	if err != nil {
		return fmt.Errorf("cannot parse `%s`: %w", name, err)
	}
	return tpl.Execute(output, data)
}

// verifyYamlLoads tries to open a YAML file and parses its content in order to
// verify that the generated file doesn't have any syntax errors
func verifyYamlLoads(fName string) error {
	contents, err := os.ReadFile(fName)
	if err != nil {
		return fmt.Errorf("cannot read file `%s`: %w", fName, err)
	}
	var data struct{}
	if err := yaml.Unmarshal(contents, &data); err != nil {
		return fmt.Errorf("cannot parse YAML file: %w", err)
	}
	return nil
}
