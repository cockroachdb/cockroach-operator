package resource_test

import (
	"bytes"
	"flag"
	"fmt"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/labels"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"path/filepath"

	"testing"
)

var update = flag.Bool("update", false, "update the golden files of this test")

func TestStatefulSetBuilder(t *testing.T) {
	sc := testutil.InitScheme(t)

	decoder, encoder := testutil.Yamlizers(t, sc)

	inputSuffix := "_in.yaml"
	folder := filepath.Join("testdata", filepath.FromSlash(t.Name()))
	testInputs := filepath.Join(folder, "/*"+inputSuffix)
	clusterFiles, err := filepath.Glob(testInputs)
	if err != nil || len(clusterFiles) == 0 {
		t.Fatalf("failed to find cluster specs %s: %v", testInputs, err)
	}

	for _, inFile := range clusterFiles {
		testName := inFile[len(folder)+1 : len(inFile)-len(inputSuffix)]
		clusterObj := decoder(load(t, inFile))
		cr, ok := clusterObj.(*api.CrdbCluster)
		if !ok {
			t.Fatal("failed to deserialize CrdbCluster")
		}
		commonLabels := labels.Common(cr)
		cluster := resource.NewCluster(cr)

		t.Run(testName, func(t *testing.T) {
			actual, err := resource.StatefulSetBuilder{
				Cluster:         &cluster,
				StatefulSetName: "test-cluster",
				Nodes:           cr.Spec.Nodes,
				Selector:        commonLabels.Selector(),
				NodeSelector: map[string]string{
					"failure-domain.beta.kubernetes.io/zone": "zone-a",
				},
				JoinStr:  "test-cluster-0.test-cluster.test-ns:26257",
				Locality: "",
			}.Build()
			require.NoError(t, err)

			var buf bytes.Buffer
			err = encoder(actual, &buf)
			require.NoError(t, err)

			expectedStr := testutil.ReadOrUpdateGoldenFile(t, buf.String(), *update)
			expected := decoder([]byte(expectedStr))

			diff := cmp.Diff(expected, actual, testutil.RuntimeObjCmpOpts...)
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}

func load(t *testing.T, file string) []byte {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("failed to load yaml file %s: %v", file, err)
	}

	return content
}
