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

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
)

func InitScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Errorf("failed to initialize Kubernetes scheme: %v", err)
	}
	if err := api.AddToScheme(scheme); err != nil {
		t.Errorf("failed to initialize CRDB scheme: %v", err)
	}

	return scheme
}

func ReadOrUpdateGoldenFile(t *testing.T, content string, update bool) string {
	t.Helper()

	gf := filepath.Join("testdata", filepath.FromSlash(t.Name())+".golden")
	if update {
		if err := ioutil.WriteFile(gf, []byte(content), 0644); err != nil {
			t.Fatalf("failed to update golden file %s: %v", gf, err)
		}
	}

	g, err := ioutil.ReadFile(gf)
	if err != nil {
		t.Fatalf("failed to read goldenfile %s: %v", gf, err)
	}

	return string(g)
}

func Yamlizers(t *testing.T, scheme *runtime.Scheme) (func([]byte) runtime.Object, func(runtime.Object, io.Writer) error) {
	codecs := serializer.NewCodecFactory(scheme)

	decode := codecs.UniversalDeserializer().Decode

	yaml, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), "application/yaml")
	if !ok {
		t.Fatalf("no yaml encoder")
	}

	encoder := codecs.EncoderForVersion(yaml.Serializer, alwaysFirstKind{})

	return func(b []byte) runtime.Object {
		obj, kind, err := decode(b, nil, nil)
		if err != nil {
			t.Fatalf("error decoding %v: %v", kind, err)
		}
		return obj
	}, encoder.Encode
}

type alwaysFirstKind struct{}

func (k alwaysFirstKind) Identifier() string {
	return "fake"
}

func (alwaysFirstKind) KindForGroupVersionKinds(kinds []schema.GroupVersionKind) (target schema.GroupVersionKind, ok bool) {
	return kinds[0], true
}
