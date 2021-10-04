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
	"reflect"
	"testing"

	//"github.com/Masterminds/semver/v3"
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
		if isValid(tc.version) != tc.valid {
			t.Errorf("expected %t for valid(`%s`) ", tc.valid, tc.version)
		}
	}
}

func TestSortVersions(t *testing.T) {
	versions := []string{"v1.2.3", "v2.10.5", "v2.2.3", "v2.1.1", "v1.11.0"}
	expected := []string{"v1.2.3", "v1.11.0", "v2.1.1", "v2.2.3", "v2.10.5"}
	got := sortVersions(versions)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected `%s`, got `%s`", expected, got)
	}
}
