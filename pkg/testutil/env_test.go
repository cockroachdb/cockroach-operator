/*
Copyright 2026 The Cockroach Authors

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

package testutil_test

import (
	"os"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestWithEnv(t *testing.T) {
	user := os.Getenv("USER")
	require.NotEmpty(t, user)

	_, ok := os.LookupEnv("WITH_ENV_TEST")
	require.False(t, ok)

	// inside function, they should be set appropriately
	WithEnv(map[string]string{"USER": "testy", "WITH_ENV_TEST": "true"}, func() {
		require.Equal(t, "testy", os.Getenv("USER"))
		require.Equal(t, "true", os.Getenv("WITH_ENV_TEST"))
	})

	// back as it was
	require.Equal(t, user, os.Getenv("USER"))
	_, ok = os.LookupEnv("WITH_ENV_TEST")
	require.False(t, ok)
}
