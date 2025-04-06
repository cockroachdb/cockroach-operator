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

package kuberecord

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/dnaeon/go-vcr/recorder"
)

const (
	// ModeReplaying will return a new *rest.Config that will instead generate HTTP responses from a YAML file on disk. No
	// network calls will be made in this mode.
	ModeReplaying = recorder.ModeReplaying
	// ModeRecording will inject fingerprinted headers into all requests and record the resulting request and response to
	// disk for future playback.
	ModeRecording = recorder.ModeRecording
	// ModeDisabled will disable both recording and replaying. Recorder will No-op in this mode.
	ModeDisabled = recorder.ModeDisabled
)

var modeString = flag.String(
	"kuberecord",
	"replay",
	"sets the global mode for kuberecord; valid values are disabled, record, and replay",
)

// Mode returns the globally configured recordMode for kuberecord based on the -kuberecord CLI flag and the KUBERECORD
// environment variable. If an unexpected value is encountered, Mode will panic.
//
// Valid values are "", record, replay, disable, and disabled.
func Mode() recorder.Mode {
	// Use environment variable if it exists, else fall back to flag.
	mode, ok := os.LookupEnv("KUBERECORD")
	if !ok {
		mode = *modeString
	}

	switch mode {
	case "record", "":
		return ModeRecording
	case "replay":
		return ModeReplaying
	case "disabled", "disable":
		return ModeDisabled
	default:
		panic(fmt.Sprintf("unknown mode: %s", mode))
	}
}

// IsEnabled is a helper method that returns true if Mode() is not disabled.
func IsEnabled() bool {
	return Mode() != ModeDisabled
}

// MaxBackOffInterval is a convenience method that returns the amount of time that backoff logic in tests should sleep
// between polling attempts (e.g. to see if a K8s resource is ready). In ModeReplaying, this function returns 0, so that
// replays will run as fast as possible. Otherwise, it returns the specified default value.
func MaxBackOffInterval(defaultVal time.Duration) time.Duration {
	if Mode() == ModeReplaying {
		return 0
	}
	return defaultVal
}

// BackOff is a convenience method meant to be used with kube.Options.BackOff.  In ModeReplaying, a backoff function
// that does not ever sleep is returned.  Otherwise, nil is returned, indicating that the default backoff option should
// be used.
func BackOff() func() backoff.BackOff {
	if Mode() == ModeReplaying {
		return zeroBackOffFunc
	}
	return nil
}

// backoff.ZeroBackOff is thread-safe, so return global instance.
var zeroBackOff = &backoff.ZeroBackOff{}
var zeroBackOffFunc = func() backoff.BackOff {
	return zeroBackOff
}
