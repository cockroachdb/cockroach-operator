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

// Package kuberecord provides record and playback functionality for testing Kubernetes calls. Think VCR, but
// specifically for K8s.
package kuberecord

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

const (
	// when replaying, bump rate limit params
	replayQueriesPerSecond = 10000
	replayBurst            = 100000000
)

// Option defines a an option for running a Recorder.
type Option interface {
	apply(*options)
}

type options struct {
	name        string
	mode        recorder.Mode
	cassetteDir string
}

type optFunc func(*options)

func (f optFunc) apply(o *options) { f(o) }

// WithName sets the recording name for the cassette.
func WithName(name string) Option {
	return optFunc(func(o *options) { o.name = name })
}

// WithCassetteDir sets the directory in which to store recorded interactions.
func WithCassetteDir(path string) Option {
	return optFunc(func(o *options) { o.cassetteDir = path })
}

// WithMode sets the playback mode to use (e.g. ModeReplaying).
func WithMode(mode recorder.Mode) Option {
	return optFunc(func(o *options) { o.mode = mode })
}

// Recorder wraps and returns a *rest.Config according the mode returned by Mode.
//
// In ModeDisabled, restCfg is returned unmodified. In ModeRecording, all HTTP traffic will be recorded and written to
// a YAML file when the test is cleaned up. In ModeReplaying, a YAML file will be loaded from disk and all responses
// will be served from there. If a request is not found the test will panic.
//
// YAML files are (by default) written/read to/from testdata/<test name>.vcr.yaml.
//
// In record mode, the recorder hashes the body of every HTTP request and adds a header with the checksum. In replay
// mode, the recorder looks up the appropriate response to serve using that checksum.
//
// All recorded interactions must be deterministic in order for playback mode to work. For tests that use random values,
// consider setting a constant seed.  For more complicated situations like key generation, consider writing a
// deterministic implementation of the relevant interfaces. If one or more requests is removed from an interaction, a
// new recording will not need to be generated. Failures will only occur when unexpected requests occur.
//
// Recorder is not a replacement for test assertions. It is a tool that allows Kubernetes tests run quickly within a
// unittest framework. Test should still make assertions about the end state of a cluster and the healthiness of the
// services running in it. If a cluster is going to be reused in a test, it is recommended to make assertions about the
// environment ahead before the test runs. For example, assert that the namespace the test is running does not exist
// before it is created to ensure no previous data is leaked. Replayed interactions will playback just fine but this
// will ensure that future recordings don't run on an unclean environment.
//
// Most test cases should follow the following template:
//       var restCfg *rest.Config
//       if kuberecord.Mode() != kuberecorder.ModeReplaying {
//               restCfg = LoadOrCreateARealRestConfig()
//       }
//       restCfg = Recorder(t, restCfg)
//       // Make initial  assertions about the environment
//       AssertNamespaceDoesntExist(t, restCfg)
//       // Run your tests and assertions
//       // Cleanup the cluster, ideally back to the initial state.
func Recorder(t *testing.T, cfg *rest.Config, opts ...Option) *rest.Config {
	o := &options{
		name:        t.Name(),
		mode:        Mode(),
		cassetteDir: "testdata",
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	// no-op if kuberecord is disabled
	if o.mode == ModeDisabled {
		return cfg
	}

	return wrapConfig(t, cloneConfig(cfg, o.mode), o)
}

func cloneConfig(cfg *rest.Config, mode recorder.Mode) *rest.Config {
	if mode == ModeReplaying {
		// When replaying disable any rate limiting that's built into the rest.Config to ensure tests run as fast as
		// possible.  Don't inherit any values from the original restCfg as we don't need them and it might be nil or empty.
		return &rest.Config{QPS: replayQueriesPerSecond, Burst: replayBurst}
	}

	// Config.Wrap modifies the Config.WrapTransport field, so make a shallow clone.
	clone := *cfg
	return &clone
}

func wrapConfig(t *testing.T, cfg *rest.Config, o *options) *rest.Config {
	// Share the same mutex and underlying recorder no matter how many times the returned rest.Config instance is used
	// (e.g. kube.Ctl vs. ClientSet). Each time the rest.Config is used to establish a new K8s connection, Wrap will be
	// called. All requests should be recorded into a single file, regardless of which of those K8s connections is used.
	var mut sync.Mutex
	var rec *recorder.Recorder

	cfg.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		mut.Lock()
		defer mut.Unlock()

		// Initialize our recorder if it has not yet been initialized.
		var err error
		if rec == nil {
			rec, err = recorder.NewAsMode(filepath.Join(o.cassetteDir, o.name+".vcr"), o.mode, rt)
			require.NoError(t, err)

			t.Cleanup(func() { require.NoError(t, rec.Stop()) })

			rec.AddFilter(Filter)
			rec.SetMatcher(Matcher)
		}

		// Ensure that all our requests will be stamped with a checksum for match _before_ they make it into the recorder.
		return &kubeRecorder{mut: &mut, rec: rec, rt: rt}
	})

	return cfg
}

type kubeRecorder struct {
	mut *sync.Mutex
	rec *recorder.Recorder
	rt  http.RoundTripper
}

func (f *kubeRecorder) RoundTrip(r *http.Request) (*http.Response, error) {
	f.mut.Lock()
	defer f.mut.Unlock()

	// Swap in the recorder's underlying transport, since different kubeRecorder instances can have different transports,
	// even though all of them use the same recorder.
	f.rec.SetTransport(f.rt)

	var body []byte
	hasher := sha1.New()

	var err error
	if r.Body != nil {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		// Reset the body to something usable
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		if _, err := hasher.Write(body); err != nil {
			// "It never returns an error."
			// - https://golang.org/pkg/hash/#Hash
			panic(err)
		}
	}

	r.Header[fingerPrintHeader] = []string{fmt.Sprintf("%x", hasher.Sum(nil))}

	resp, err := f.rec.RoundTrip(r)

	if errors.Is(err, cassette.ErrInteractionNotFound) {
		panic(fmt.Sprintf(
			`Requested interaction not found.\nMethod: %s\nURL: %s\nBody: %s
Do you need to re-record this interaction?
Pass -kuberecord='' to your tests and -update if Golden files need to be updated.
`,
			r.Method,
			r.URL,
			body,
		))
	}

	return resp, err
}
