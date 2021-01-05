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
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
)

func TestBoilerPlates(t *testing.T) {

	project := os.Getenv("TEST_WORKSPACE")
	if project == "" {
		t.Fatal("unable to read TEST_WORKSPACE environment variable")
	}

	runFiles := os.Getenv("RUNFILES_DIR")
	if runFiles == "" {
		t.Fatal("unable to read RUNFILES_DIR environment variable")
	}

	rootDir := fmt.Sprintf("%s/%s", runFiles, project)
	bpDir := fmt.Sprintf("%s/hack/boilerplate", rootDir)

	v := testutil.NewValidateHeaders(nil, rootDir, bpDir, "")
	nonValidFiles, err := v.Validate()
	if err != nil {
		t.Fatal("error running validate", err)
	}

	if nonValidFiles != nil {
		t.Log("The following files have invalid headers")
		for _, filename := range *nonValidFiles {
			t.Log(strings.Replace(filename, rootDir+"/", "", 1))
		}
		t.Fatal("Please update files with correct header")
	}

}
