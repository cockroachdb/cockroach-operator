/*
Copyright 2023 The Cockroach Authors

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

package update

import (
	"testing"

	semver "github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestPerserveMatches(t *testing.T) {
	tests := []struct {
		description     string
		wantVersion     *semver.Version
		currentVersion  *semver.Version
		perserveVersion *semver.Version
		result          bool
	}{
		{
			"perserved is not set correctly",
			semver.MustParse("20.1.6"),
			semver.MustParse("21.1.5"),
			semver.MustParse("20.2.5"),
			false,
		},
		{
			"perserved is not set correctly jumping two versions",
			semver.MustParse("19.1.6"),
			semver.MustParse("21.1.5"),
			semver.MustParse("20.2.5"),
			false,
		},
		{
			"perserved is set correctly, exact version",
			semver.MustParse("20.1.6"),
			semver.MustParse("21.1.5"),
			semver.MustParse("20.1.6"),
			true,
		},
		{
			"perserved is set correctly, different patch version",
			semver.MustParse("20.1.14"),
			semver.MustParse("21.1.5"),
			semver.MustParse("20.1.6"),
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			_, err := checkDowngradeAllowed(test.wantVersion, test.currentVersion, test.perserveVersion)
			if test.result {
				require.Nil(t, err, "test should have passed, but an error was returned")
			} else {
				require.Error(t, err, "test should have failed, and no error was returned")
			}
		})
	}
}
