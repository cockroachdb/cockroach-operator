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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	versionRegxp = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
)

// Step defines an action to be taken during the release process
type Step interface {
	Apply(version string) error
}

// StepFn is a function that implements Step.
type StepFn func(version string) error

// Apply applies the function.
func (fn StepFn) Apply(version string) error {
	return fn(version)
}

// CmdFn describes a function that runs a Cmd.
type CmdFn func(cmd *exec.Cmd) error

// ExecFn describes a function that executes shell commands.
type ExecFn func(cmd string, args, env []string) error

// FileFn describes a function that reads a file and returns it's contents
type FileFn func(path string) ([]byte, error)

// ValidateVersion ensures the supplied version matches our expected version regexp.
func ValidateVersion() Step {
	return StepFn(func(version string) error {
		if !versionRegxp.MatchString(version) {
			return fmt.Errorf("invalid version '%s'. Must be of the form N.N.N", version)
		}

		return nil
	})
}

// EnsureUniqueVersion verifies that this is a new version by checking the existing tags.
func EnsureUniqueVersion(fn CmdFn) Step {
	return StepFn(func(version string) error {
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		cmd := exec.Command("git", "tag")
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := fn(cmd); err != nil {
			return fmt.Errorf("failed to get tags: %s - %s", stderr.String(), err)
		}

		for _, v := range strings.Split(stdout.String(), "\n") {
			if v == "" {
				continue
			}

			// tags have a `v` prefix, so we remove it for comparison
			if v[1:] == version {
				return fmt.Errorf("version already exists")
			}
		}

		return nil
	})
}

// UpdateVersion sets the version in version.txt
func UpdateVersion() Step {
	return StepFn(func(version string) error {
		// setting the mode to 0644 to match the existing permissions: r/w for current user, read-only for everyone else.
		return ioutil.WriteFile("version.txt", []byte(version), 0644)
	})
}

// CreateReleaseBranch creates a new branch for the release named release-<version>.
func CreateReleaseBranch(fn ExecFn) Step {
	return StepFn(func(version string) error {
		return fn(
			"git",
			[]string{"checkout", "-b", fmt.Sprintf("release-%s", version)},
			os.Environ(),
		)
	})
}

// GenerateFiles runs make release/gen-files passing the appropriate channel options based on the version.
func GenerateFiles(fn ExecFn) Step {
	return StepFn(func(version string) error {
		ch := "stable"
		defaultCh := "stable"

		return fn(
			"make",
			[]string{"release/gen-files", "CHANNELS=" + ch, "DEFAULT_CHANNEL=" + defaultCh},
			os.Environ(),
		)
	})
}

// UpdateChangelog ensures that the release is setup correctly in the changelog and that a new [Unreleased] section is
// added appropriately.
func UpdateChangelog(fn FileFn) Step {
	const fileName = "CHANGELOG.md"
	const urlFmt = "https://github.com/cockroachdb/cockroach-operator/compare/v%s...%s"

	return StepFn(func(version string) error {
		data, err := fn(fileName)
		if err != nil {
			return err
		}

		// get the existing and new [Unreleased] lines
		start := bytes.Index(data, []byte("[Unreleased]"))
		end := bytes.Index(data[start:], []byte("\n"))
		prevUnreleased := data[start : start+end]
		newUnreleased := []byte(fmt.Sprintf("[Unreleased](%s)", fmt.Sprintf(urlFmt, version, "master")))

		// fix up the previous unreleased line to reference the new version
		latestRelease := bytes.Replace(prevUnreleased, []byte("...master"), []byte("...v"+version), 1)
		latestRelease = bytes.Replace(latestRelease, []byte("[Unreleased]"), []byte(fmt.Sprintf("# [v%s]", version)), 1)

		// update to include the new and previous versions
		newUnreleased = append(newUnreleased, append([]byte("\n\n"), latestRelease...)...)
		data = bytes.Replace(data, prevUnreleased, newUnreleased, 1)

		return ioutil.WriteFile(fileName, data, 0644)
	})
}
