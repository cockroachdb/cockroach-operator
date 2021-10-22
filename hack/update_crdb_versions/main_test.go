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

package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/hack/update_crdb_versions"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"v1.2.3", true},
		{"latest", false},
		{"v1.2.3-ubi", false},
		{"v19.1.5", false},
		{"v21.1.8", false},
		{"v21.1.9", true},
		{"v20.2.11", true},
	}

	for _, tc := range tests {
		if IsValid(tc.version) != tc.valid {
			t.Errorf("expected %t for valid(`%s`) ", tc.valid, tc.version)
		}
	}
}

func TestSortVersions(t *testing.T) {
	versions := []string{"v1.2.3", "v2.10.5", "v2.2.3", "v2.1.1", "v1.11.0"}
	expected := []string{"v1.2.3", "v1.11.0", "v2.1.1", "v2.2.3", "v2.10.5"}
	got := SortVersions(versions)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected `%s`, got `%s`", expected, got)
	}
}

func TestGetVersions(t *testing.T) {
	expected := []string{"v1.2.3", "v1.2.3+test.01"}

	data := `{"data":[{"repositories": [{"tags": [{"name": "v1.2.3"}, {"name": "v1.2.3+test.01"}]}]}]}`
	resp := CrdbVersionsResponse{}
	json.Unmarshal([]byte(data), &resp)
	got := GetVersions(resp)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected `%s`, got `%s`", expected, got)
	}
}

func TestCrdbVersionsFile(t *testing.T) {
	versions := []string{"v1.2.3", "v1.2.3+test.01"}

	output := `CrdbVersions:
- v1.2.3
- v1.2.3+test.01
`
	expected := append([]byte(CrdbVersionsFileDescription), []byte(output)...)

	tmpdir := t.TempDir()
	filePath := path.Join(tmpdir, CrdbVersionsFileName)
	err := GenerateCrdbVersionsFile(versions, filePath)
	if err != nil {
		t.Fatalf("error generating file: %s", err)
	}

	got, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("error reading generated file: %s", err)
	}

	if bytes.Compare(got, expected) != 0 {
		t.Errorf("Expected `%s`, got `%s`", expected, got)
	}
}
