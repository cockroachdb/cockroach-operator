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

package env_test

import (
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
)

func TestCreateEnv(t *testing.T) {
	if env := CreateActiveEnvForTest(); env == nil {
		t.Log("env is nil")
		t.Fail()
	}
}
