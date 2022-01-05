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
	"bytes"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestIsStable(t *testing.T) {
	tests := []struct {
		version string
		stable  bool
	}{
		{"1.2.3", true},
		{"1.2.3+test.01", false},
		{"1.2.3-alpha.-1", false},
		{"1.2.0-x.Y.0+metadata", false},
		{"1.2.0-x.Y.0+metadata-width-hypen", false},
		{"1.2.3-rc1-with-hypen", false},
		{"1.2.2147483648", true},
		{"1.2147483648.3", true},
		{"2147483648.3.0", true},
	}

	for _, tc := range tests {
		v, err := semver.NewVersion(tc.version)
		if err != nil {
			t.Fatalf("error parsing %q: %s", tc.version, err)
		}
		if isStable(*v) != tc.stable {
			t.Errorf("expected %t for isStable(`%s`) ", tc.stable, v)
		}
	}
}

func TestDotsToUndescore(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{"1.2.3", "1_2_3"},
		{"1.2.3+test.01", "1_2_3+test_01"},
		{"1.2.3-alpha.-1", "1_2_3-alpha_-1"},
	}
	for _, tc := range tests {
		got := dotsToUnderscore(tc.version)
		if got != tc.expected {
			t.Errorf("expected %q for dotsToUnderscore(`%s`), got %s", tc.expected, tc.version, got)
		}
	}
}

func TestReadCrdbVersions(t *testing.T) {
	s := `
CrdbVersions:
  - v21.2.0-beta.1
  - v21.1.0
  - v21.1.1
  - v20.1.11`
	versions, err := readCrdbVersions(strings.NewReader(s))
	if err != nil {
		t.Fatalf("cannot read versions file: %s", err)
	}
	expected := "v20.1.11, v21.1.0, v21.1.1, v21.2.0-beta.1, "
	got := ""
	for _, v := range versions {
		got += v.Original() + ", "
	}
	if got != expected {
		t.Errorf("Expected `%s`, got `%s`", expected, got)
	}
}

func TestGenerateTemplateData(t *testing.T) {
	versions := []string{"1.2.3", "1.2.3+test.01", "1.2.3-alpha.-1"}
	var crdbVersions []*semver.Version
	for _, r := range versions {
		v, err := semver.NewVersion(r)
		if err != nil {
			t.Fatalf("error parsing %q: %s", r, err)
		}
		crdbVersions = append(crdbVersions, v)
	}
	data, err := generateTemplateData(crdbVersions, "2.0.1")
	if err != nil {
		t.Fatalf("error generating data: %s", err)
	}
	if len(data.CrdbVersions) != len(crdbVersions) {
		t.Error("CrdbVersions len mismatch")
	}
	if data.LatestStableCrdbVersion != "1.2.3" {
		t.Errorf("Expected LatestStableCrdbVersion 1.2.3, got %s", data.LatestStableCrdbVersion)
	}
	if data.OperatorVersion != "2.0.1" {
		t.Errorf("Expected OperatorVersion 2.0.1, got %s", data.OperatorVersion)
	}
}

func TestGenerateFile(t *testing.T) {
	versions := []string{"1.2.3", "1.2.3+test.01", "1.2.3-alpha.-1"}
	var data templateData
	for _, r := range versions {
		v, err := semver.NewVersion(r)
		if err != nil {
			t.Fatalf("error parsing %q: %s", r, err)
		}
		data.CrdbVersions = append(data.CrdbVersions, v)
	}
	// test custom function
	tplText := "{{range .CrdbVersions}}{{if stable .}}{{ underscore .Original }} {{end}}{{end}}{{range .CrdbVersions}}{{underscore .Original}} {{ end }}"
	expected := "1_2_3 1_2_3 1_2_3+test_01 1_2_3-alpha_-1 "
	var output bytes.Buffer
	err := generateFile("name", tplText, &output, data)
	if err != nil {
		t.Fatalf("error generating template")
	}
	if output.String() != expected {
		t.Errorf("Expected `%s`, got `%s`", expected, output.String())
	}
}
