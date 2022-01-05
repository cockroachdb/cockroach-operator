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
	"context"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/kuberecord"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestRecorder(t *testing.T) {
	env := &envtest.Environment{BinaryAssetsDirectory: env.ExpandPath("hack", "bin")}

	cfg, err := env.Start()
	require.NoError(t, err)
	defer func(env *envtest.Environment) {
		if err := env.Stop(); err != nil {
			t.Log(err.Error())
		}
	}(env)

	cfg = Recorder(
		t,
		cfg,
		WithName("kuberecord_demo"),
		WithCassetteDir("testdata/cassettes"),
	)

	c, err := client.New(cfg, client.Options{})
	require.NoError(t, err)
	require.NotNil(t, c)

	ns := &v1.Namespace{ObjectMeta: meta.ObjectMeta{Name: "test-ns"}}
	require.NoError(t, c.Create(context.TODO(), ns))
}
