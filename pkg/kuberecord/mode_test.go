/*
Copyright 2022 The Cockroach Authors

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

package kuberecord_test

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	. "github.com/cockroachdb/cockroach-operator/pkg/kuberecord"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/stretchr/testify/require"
)

func TestMode(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		env  string
		exp  recorder.Mode
	}{
		{name: "neither flag nor env set", exp: ModeReplaying},
		{name: "flag is set to record", arg: "record", exp: ModeRecording},
		{name: "flag is set to replay", arg: "replay", exp: ModeReplaying},
		{name: "flag is set to disabled", arg: "disabled", exp: ModeDisabled},
		{name: "env is set to record", env: "record", exp: ModeRecording},
		{name: "env is set to replay", env: "replay", exp: ModeReplaying},
		{name: "env is set to disabled", env: "disabled", exp: ModeDisabled},
		{name: "both are set", arg: "record", env: "replay", exp: ModeReplaying},
	}

	for _, tt := range tests {
		env := map[string]string{}
		if tt.env != "" {
			env["KUBERECORD"] = tt.env
		}

		testutil.WithEnv(env, func() {
			if tt.arg != "" {
				defer setFlag(tt.arg)()
			}

			require.Equal(t, tt.exp, Mode(), tt.name)
		})
	}

	t.Run("invalid values", func(t *testing.T) {
		require.Panics(t, func() {
			testutil.WithEnv(map[string]string{"KUBERECORD": "unknown"}, func() {
				Mode()
			})
		})
	})
}

func TestIsEnabled(t *testing.T) {
	tests := map[string]bool{
		"record":   true,
		"replay":   true,
		"disabled": false,
	}

	for env, res := range tests {
		testutil.WithEnv(map[string]string{"KUBERECORD": env}, func() {
			require.Equal(t, res, IsEnabled())
		})
	}
}

func TestMaxBackOffInterval(t *testing.T) {
	tests := map[string]time.Duration{
		"record":   10,
		"replay":   0,
		"disabled": 10,
	}

	for env, dur := range tests {
		testutil.WithEnv(map[string]string{"KUBERECORD": env}, func() {
			require.Equal(t, dur, MaxBackOffInterval(10), env)
		})
	}
}

func TestBackoff(t *testing.T) {
	tests := map[string]backoff.BackOff{
		"record":   nil,
		"replay":   new(backoff.ZeroBackOff),
		"disabled": nil,
	}

	for env, expT := range tests {
		testutil.WithEnv(map[string]string{"KUBERECORD": env}, func() {
			fn := BackOff()
			if expT == nil {
				require.Nil(t, fn)
				return
			}

			require.IsType(t, expT, fn())
		})
	}
}

func setFlag(v string) func() {
	oldArgs := os.Args
	os.Args = append(os.Args, "-kuberecord", v)
	flag.Parse()

	return func() {
		os.Args = oldArgs
		flag.Parse()
	}
}
