/*
Copyright 2025 The Cockroach Authors

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
	"reflect"
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
		{
			"is skipping innovative release",
			semver.MustParse("24.3"),
			semver.MustParse("24.1"),
			true,
		},
		{
			"is skipping innovative and regular release",
			semver.MustParse("25.1"),
			semver.MustParse("24.1"),
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
		{
			"is skipping innovative release",
			semver.MustParse("24.1"),
			semver.MustParse("24.3"),
			true,
		},
		{
			"is skipping innovative and regular release",
			semver.MustParse("24.1"),
			semver.MustParse("25.1"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, isBackOneMajorVersion(test.wantVersion, test.currentVersion) == test.result, "isBackwardOneMajorVersion test failed")
		})
	}
}

func TestGetNextReleases(t *testing.T) {
	tests := []struct {
		description    string
		currentVersion string
		result         []string
	}{
		{
			"returns the possible upgrade targets for 24.1",
			"24.1",
			[]string{"24.2", "24.3"},
		},
		{
			"returns the possible upgrade targets for 24.3",
			"24.3",
			[]string{"25.1", "25.2"},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, reflect.DeepEqual(getNextReleases(test.currentVersion), test.result),
				"getNextReleases test failed")
		})
	}
}

func TestGetPreviousReleases(t *testing.T) {
	tests := []struct {
		description    string
		currentVersion string
		result         []string
	}{
		{
			"returns the possible rollback targets for 24.3",
			"24.3",
			[]string{"24.2", "24.1"},
		},
		{
			"returns the possible upgrade targets for 25.2",
			"25.2",
			[]string{"25.1", "24.3"},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, reflect.DeepEqual(getPreviousReleases(test.currentVersion), test.result),
				"getPreviousReleases test failed")
		})
	}
}

func TestGenerateReleases(t *testing.T) {
	tests := []struct {
		description string
		uptoYear    int
		result      []string
	}{
		{
			"returns possible releases till 2025",
			25,
			[]string{"24.1", "24.2", "24.3", "25.1", "25.2", "25.3", "25.4"},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, reflect.DeepEqual(generateReleases(test.uptoYear), test.result), "generateReleases test failed")
		})
	}
}

func TestGetReleaseType(t *testing.T) {
	tests := []struct {
		description  string
		major, minor int
		result       ReleaseType
	}{
		{
			"returns releaseType of a 24.2",
			24, 2,
			Innovative,
		},
		{
			"returns releaseType of a 25.1",
			24, 2,
			Innovative,
		},
		{
			"returns releaseType of a 24.3",
			24, 3,
			Regular,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			require.True(t, reflect.DeepEqual(getReleaseType(test.major, test.minor), test.result),
				"getReleaseType test failed")
		})
	}
}
