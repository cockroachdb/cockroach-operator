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
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: versionbump major|minor|patch version")
		os.Exit(1)
	}
	bumpType, rawVersion := os.Args[1], os.Args[2]
	version, err := semver.NewVersion(rawVersion)
	if err != nil {
		fmt.Printf("Cannot convert version `%s`: %s\n", rawVersion, err)
		os.Exit(1)
	}
	var newVerson semver.Version
	switch bumpType {
	case "major":
		newVerson = version.IncMajor()
	case "minor":
		newVerson = version.IncMinor()
	case "patch":
		newVerson = version.IncPatch()

	}
	fmt.Println(newVerson.Original())
}
