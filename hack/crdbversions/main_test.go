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

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
		require.Equal(t, tc.expected, dotsToUnderscore(tc.version))
	}
}

func TestReadCrdbVersions(t *testing.T) {
	s := `
CrdbVersions:
  - tag: v21.2.0
  - tag: v21.1.0
  - tag: v21.1.1
  - tag: v20.1.11`
	versions, err := readCrdbVersions(strings.NewReader(s))
	require.NoError(t, err)

	expected := []string{"v20.1.11", "v21.1.0", "v21.1.1", "v21.2.0"}
	got := make([]string, len(expected))
	for idx, v := range versions {
		got[idx] = v.Tag
	}

	require.ElementsMatch(t, expected, got)
}

func TestGenerateTemplateData(t *testing.T) {
	versions := []string{"1.2.1", "1.2.2", "1.2.3"}
	var crdbVersions []crdbVersion
	for _, r := range versions {
		crdbVersions = append(crdbVersions, crdbVersion{Tag: r})
	}

	data, err := generateTemplateData(crdbVersions, "2.0.1", "cockroachdb/cockroach-operator")
	require.NoError(t, err)
	require.Len(t, data.CrdbVersions, len(crdbVersions))
	require.Equal(t, "1.2.3", data.LatestStableCrdbVersion)
	require.Equal(t, "2.0.1", data.OperatorVersion)
}

func TestGenerateFile(t *testing.T) {
	versions := []string{"1.2.3", "1.2.3+test.01", "1.2.3-alpha.-1"}
	var data templateData
	for _, r := range versions {
		data.CrdbVersions = append(data.CrdbVersions, crdbVersion{Tag: r})
	}

	// test custom function
	tplText := "{{range .CrdbVersions}}{{ underscore .Tag }} {{end}}"
	expected := "1_2_3 1_2_3+test_01 1_2_3-alpha_-1 "

	var output bytes.Buffer
	require.NoError(t, generateFile("name", tplText, &output, data))
	require.Equal(t, expected, output.String())
}
