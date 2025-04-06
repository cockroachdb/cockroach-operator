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

package testutil

import "testing"

// Step is used to build a test
type Step struct {
	Name string
	Test func(t *testing.T)
}

// Steps
type Steps []Step

// WithSteps appends Steps
func (ss Steps) WithStep(s Step) Steps {
	return append(ss, s)
}

// Run execs the various Steps that describe Tests
func (ss Steps) Run(t *testing.T) {
	for _, s := range ss {
		if !t.Run(s.Name, s.Test) {
			t.FailNow()
		}
	}
}
