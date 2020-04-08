package testutil

import (
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	"testing"

	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
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
