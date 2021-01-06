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

package update

import (
	"testing"

	semver "github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/require"
)

func TestIsPatch(t *testing.T) {
	tests := []struct {
		description    string
		wantVersion    *semver.Version
		currentVersion *semver.Version
		result         bool
	}{
		{
			"isPatch true",
			semver.MustParse("20.1.6"),
			semver.MustParse("20.1.5"),
			true,
		},
		{
			"isPatch false",
			semver.MustParse("19.2.6"),
			semver.MustParse("20.1.5"),
			false,
		},
		{
			"isPatch false test two",
			semver.MustParse("19.2.6"),
			semver.MustParse("19.3.5"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, isPatch(test.wantVersion, test.currentVersion) == test.result, "patch result failed")
		})
	}

}

func TestIsForwardOneMajorVersion(t *testing.T) {
	tests := []struct {
		description    string
		wantVersion    *semver.Version
		currentVersion *semver.Version
		result         bool
	}{
		{
			"forward minor update",
			semver.MustParse("20.1.6"),
			semver.MustParse("20.1.5"),
			false,
		},
		{
			"forward major update",
			semver.MustParse("20.1.6"),
			semver.MustParse("19.2.5"),
			true,
		},
		{
			"backward minor version",
			semver.MustParse("19.2.6"),
			semver.MustParse("19.3.5"),
			false,
		},
		{
			"is backward major version",
			semver.MustParse("19.2.6"),
			semver.MustParse("21.3.5"),
			false,
		},
		{
			"is two major",
			semver.MustParse("22.2.6"),
			semver.MustParse("19.3.5"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, isForwardOneMajorVersion(test.wantVersion, test.currentVersion) == test.result, "isForwardOneMajorVersion test failed")
		})
	}
}

func TestIsBackOneMajorVersion(t *testing.T) {
	tests := []struct {
		description    string
		wantVersion    *semver.Version
		currentVersion *semver.Version
		result         bool
	}{
		{
			"forward minor update",
			semver.MustParse("20.1.6"),
			semver.MustParse("20.1.5"),
			false,
		},
		{
			"backward major version 20 to 19",
			semver.MustParse("19.2.6"),
			semver.MustParse("20.1.5"),
			true,
		},
		{
			"backward major version 21 to 20",
			semver.MustParse("20.2.6"),
			semver.MustParse("21.1.5"),
			true,
		},
		{
			"is two major",
			semver.MustParse("22.2.6"),
			semver.MustParse("19.3.5"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, isBackOneMajorVersion(test.wantVersion, test.currentVersion) == test.result, "isBackwardOneMajorVersion test failed")
		})
	}
}
