package testutil

import (
	"github.com/cockroachlabs/crdb-operator/pkg/label"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"path/filepath"
	"testing"

	"github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Errorf("failed to initialize Kubernetes scheme: %v", err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Errorf("failed to initialize CRDB scheme: %v", err)
	}

	return scheme
}

func MakeStandAloneCluster(_ *testing.T) *v1alpha1.CrdbCluster {
	c := &v1alpha1.CrdbCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "test-ns",
			Labels:    make(map[string]string),
		},
		Spec: v1alpha1.CrdbClusterSpec{},
		Status: v1alpha1.CrdbClusterStatus{
			Version: "v19.2",
		},
	}

	c.Labels = label.MakeCommonLabels(c)

	v1alpha1.SetClusterSpecDefaults(&c.Spec)

	return c
}

func ReadOrUpdateGoldenFile(t *testing.T, content string, update bool) string {
	t.Helper()

	gf := filepath.Join("testdata", filepath.FromSlash(t.Name()) + ".golden")
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