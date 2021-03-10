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

package security

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach-operator/pkg/testutil/paths"
)

const defaultKeySize = 2048

// We use 366 days on certificate lifetimes to at least match X years,
// otherwise leap years risk putting us just under.
const defaultCALifetime = 10 * 366 * 24 * time.Hour  // ten years
const defaultCertLifetime = 5 * 366 * 24 * time.Hour // five years

// tempDir is like testutils.TempDir but avoids a circular import.
func tempDir(t *testing.T) (string, func()) {
	certsDir, err := ioutil.TempDir("", "certs_test")
	if err != nil {
		t.Fatal(err)
	}
	return certsDir, func() {
		if err := os.RemoveAll(certsDir); err != nil {
			t.Fatal(err)
		}
	}
}

func setPath(t *testing.T) {
	// We are running in bazel so set up the directory for the test binaries
	if os.Getenv("TEST_WORKSPACE") != "" {
		// TODO create a toolchain for this
		paths.MaybeSetEnv("PATH", "cockroach", "hack", "bin", "cockroach")
	} else {
		t.Fatal("TEST_WORKSPACE not defined.  Are you running me with bazel")
	}
}

func TestCreateCAPair(t *testing.T) {
	setPath(t)
	certsDir, cleanup := tempDir(t)
	defer cleanup()
	ca := filepath.Join(certsDir, "ca.key")

	err := CreateCAPair(certsDir, ca, defaultKeySize, defaultCALifetime, true, true)
	if err != nil {
		t.Error(err)
	}

	if !fileExists(filepath.Join(certsDir, "ca.crt")) {
		t.Fail()
	}

	if !fileExists(ca) {
		t.Fail()
	}
}

func TestCreateNodePair(t *testing.T) {
	setPath(t)
	certsDir, cleanup := tempDir(t)
	defer cleanup()
	ca := filepath.Join(certsDir, "ca.key")

	err := CreateCAPair(certsDir, ca, defaultKeySize, defaultCALifetime, true, true)
	if err != nil {
		t.Error(err)
	}

	if !fileExists(filepath.Join(certsDir, "ca.crt")) {
		t.Fail()
	}

	if !fileExists(ca) {
		t.Fail()
	}

	err = CreateNodePair(certsDir, ca, defaultKeySize, defaultCALifetime, true, []string{"*.foo.com", "bar.foo.com", "127.0.0.1"})
	if err != nil {
		t.Error(err)
	}

	if !fileExists(filepath.Join(certsDir, "node.crt")) {
		t.Fail()
	}

	if !fileExists(filepath.Join(certsDir, "node.key")) {
		t.Fail()
	}
}

func TestCreateClientPair(t *testing.T) {
	setPath(t)
	certsDir, cleanup := tempDir(t)
	defer cleanup()
	ca := filepath.Join(certsDir, "ca.key")

	// This is replacing some code
	u := &SQLUsername{
		U: "root",
	}
	err := CreateCAPair(certsDir, ca, defaultKeySize, defaultCALifetime, true, true)
	if err != nil {
		t.Error(err)
	}

	if !fileExists(filepath.Join(certsDir, "ca.crt")) {
		t.Fail()
	}

	if !fileExists(ca) {
		t.Fail()
	}

	err = CreateClientPair(certsDir, ca, defaultKeySize, defaultCALifetime, true, *u, false)
	if err != nil {
		t.Error(err)
	}

	if !fileExists(filepath.Join(certsDir, "client.root.crt")) {
		t.Fail()
	}

	if !fileExists(filepath.Join(certsDir, "client.root.key")) {
		t.Fail()
	}
}

// fileExists reports whether the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
