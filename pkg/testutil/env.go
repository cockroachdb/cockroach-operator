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

package testutil

import "os"

// ENV is an alias for map[string]string to make it a little nicer when supplying vars to WithEnv.
type ENV map[string]string

// WithEnv sets environment variables, runs the supplied function and puts the env back the way it was found.
func WithEnv(vars ENV, fn func()) {
	toReset := map[string]string{}
	toClear := make([]string, 0)

	for k, v := range vars {
		if oldV, ok := os.LookupEnv(k); ok {
			toReset[k] = oldV
		} else {
			toClear = append(toClear, k)
		}

		os.Setenv(k, v)
	}

	defer func() {
		for k, v := range toReset {
			os.Setenv(k, v)
		}

		for _, k := range toClear {
			os.Unsetenv(k)
		}
	}()

	fn()
}
