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

package openshift

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"os"
	"testing"

	"github.com/cenkalti/backoff"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kubetest2/pkg/process"
)

// This is the YAML used to create an operatorGroup
// in OpenShift to deploy the operator.
const operatorGroup = `apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: test
  namespace: test
spec:
  targetNamespaces:
  - test
`

// This is the YAML used to create a CatalogSource
// in OpenShift to have the operator appear in the Marketplace
const catalogSource = `apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: cockroach-test
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  displayName: Cockroach Test Demo
  image: {{.DockerRegistry}}/cockroachdb-operator-index:{{.AppVersion}}`

// This is the YAML used to create a subscription
// in OpenShift to deploy the operator.
const subscription = `apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: myoperator
  namespace: test
spec:
  channel: stable
  installPlanApproval: Automatic
  name: cockroachdb-certified
  source: cockroach-test
  sourceNamespace: openshift-marketplace
  startingCSV: cockroach-operator.{{.AppVersion}}`

// YAMLTemplate struct is used to render the
// above YAML files
type YAMLTemplate struct {
	AppVersion     string
	DockerRegistry string
}

// TestPackage is used to install the operator in an OpenShift
// cluster. It uses three different YAML files and uses the test
// namespace. This program uses oc and also installs example.yaml
// to test that the operator is functioning.
func TestPackaging(t *testing.T) {

	// Test that three env variables are set
	appVersion := os.Getenv("APP_VERSION")
	require.True(t, appVersion != "", "APP_VERSION env var not set")
	dockerRegistry := os.Getenv("DOCKER_REGISTRY")
	require.True(t, dockerRegistry != "", "DOCKER_REGISTRY env var not set")
	kubeconfig := os.Getenv("KUBECONFIG")
	require.True(t, kubeconfig != "", "KUBECONFIG evn var not set")

	// We bring in //hack/bin:oc as a data resource in the go_test target. This ensures that it's available on the PATH.
	env.PrependToPath(env.ExpandPath("hack", "bin"))

	// remove old crdb db if it still exists
	args := []string{
		"delete",
		"crdbclusters.crdb.cockroachlabs.com",
		"cockroachdb",
	}

	require.NoError(t, process.ExecJUnit("oc", args, os.Environ()))

	// TODO I should wait for the cockroach database
	// to stop here, we might have orphaned disks

	// remove old pvc's from last run
	args = []string{
		"-n", "test",
		"delete",
		"pvc",
		"-l",
		"app.kubernetes.io/name=cockroachdb",
	}

	require.NoError(t, process.ExecJUnit("oc", args, os.Environ()))

	// remove test namespace
	args = []string{
		"delete",
		"namespace",
		"test",
	}

	require.NoError(t, process.ExecJUnit("oc", args, os.Environ()))

	// remove crds
	args = []string{
		"delete",
		"crd",
		"crdbclusters.crdb.cockroachlabs.com",
	}

	require.NoError(t, process.ExecJUnit("oc", args, os.Environ()))

	// create a new project/namespace for testing
	args = []string{
		"new-project",
		"test",
	}

	// we cannot do a require.NoError on process, because oc tries to modify
	// the kubeconfig file, and with bazel you cannot do this
	require.NoError(t, process.ExecJUnit("oc", args, os.Environ()))

	// create a YAMLTemplate for the go templating
	yaml := &YAMLTemplate{
		AppVersion:     appVersion,
		DockerRegistry: dockerRegistry,
	}

	// create some YAML files
	catalogByte := parseTmpl(t, catalogSource, yaml)
	subByte := parseTmpl(t, subscription, yaml)

	catalogFilename := writeFile(t, catalogByte)
	subFilename := writeFile(t, subByte)

	var b bytes.Buffer
	b.WriteString(operatorGroup)

	operatorFilename := writeFile(t, b.Bytes())

	files := []string{catalogFilename, subFilename, operatorFilename}

	// install YAML files which install the operator
	for _, f := range files {
		fmt.Println(f)
		args := []string{"delete", "-f", f}
		require.NoError(t, process.ExecJUnit("oc", args, os.Environ()))
		args = []string{"create", "-f", f}
		require.NoError(t, process.ExecJUnit("oc", args, os.Environ()), "failed creating openshift operator files")
	}

	cs, err := createClientset()
	require.NoError(t, err, "failed to create clientset")

	// test that the operator pod is up and running
	// this test checks that all pods are running in a ns
	// but that works.
	require.NoError(t, testPods(t, cs), "failed finding operator")

	// create an example database
	args = []string{
		"create", "-f", "../../examples/example.yaml",
	}

	require.NoError(t, process.ExecJUnit("oc", args, os.Environ()), "failed creating crdb database")

	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	// wait till the sts is ready, no the code that we have actually does not fully work
	// and no other code is using it, and there is an issue open. It returns
	// before all of the pods are running.
	require.NoError(t,
		kube.WaitUntilAllStsPodsAreReady(
			context.TODO(),
			cs,
			testLog,
			"cockroachdb",
			"test",
			800*time.Second,
			10*time.Second,
		),
		"crdb database did not start",
	)

	// testing that all pods are running in the namespace, except for the job pod
	// this will check that the pods are up and going
	require.NoError(t, testPods(t, cs))

	t.Log("rejoice and celebrate as our operator and OpenShift are happy")

	// TODO: We do not do any openshift resource cleanup at this time,
	// but the test will run again if we need it to.
}

// letters are used to generate a random string of letters by randSeq
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// randSeq returns a string of n letters
func randSeq(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

// wriiteFile writes out b byte array to a random file in the
// os.TempDir and returns the name of that file.
func writeFile(t *testing.T, b []byte) string {
	fileName := fmt.Sprintf("%s-test.yaml", randSeq(10))
	fileName = filepath.Join(t.TempDir(), fileName)

	if err := os.WriteFile(fileName, b, 0644); err != nil {
		t.Fatal("Failed to write to temporary file", err)
	}

	return fileName
}

// parseTmpl uses go templates to render a string and a YAMLTemplate struct
func parseTmpl(t *testing.T, s string, yaml *YAMLTemplate) []byte {
	tmpl := template.Must(template.New("test").Parse(s))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, yaml)
	require.NoError(t, err)
	return buf.Bytes()
}

// createClientset creates a k8s client set
func createClientset() (*kubernetes.Clientset, error) {

	// TODO we should move this into a different pkg

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		return nil, errors.New("KUBECONFIG env variable is not set")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// TODO we should move this into a package, we are using this in
// multple places in the code base.

// backoffFactory is a replacable global for backoff creation. It may be
// replaced with shorter times to allow testing of Wait___ functions without
// waiting the entire default period
var backoffFactory = defaultBackoffFactory

func defaultBackoffFactory(maxTime time.Duration) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxTime
	return b
}

var defaultTime time.Duration = 5 * time.Minute

// testPods looks up Pods and does an exponetial backoff test
// that the Pod exist and is running. This func skips the verification job pod.
func testPods(t *testing.T, clientset *kubernetes.Clientset) error {

	// TODO move this to the testutil pkg
	f := func() error {
		// find a list of pods via the label
		pods, err := clientset.CoreV1().Pods("test").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			t.Log("getting pods errored out")
			return err
		}
		// if the pod list is zero no pods are running yet
		// and we throw an error.
		if len(pods.Items) == 0 {
			t.Logf("cannot find any pods")
			return errors.New("unable to find any pods")
		}

		// iterate through the pods and test that each pod
		// is ready
		for _, pod := range pods.Items {
			if strings.Contains(pod.Name, "-vcheck-") {
				continue // Skip the validate version pod
			}
			if !kube.IsPodReady(&pod) {
				msg := fmt.Sprintf("pod not ready: %s", pod.Name)
				t.Log(msg)
				return errors.New(msg)
			}
		}

		return nil
	}

	b := backoffFactory(defaultTime)
	// run the func with a backoff factory
	return backoff.Retry(f, backoff.WithContext(b, context.TODO()))
}
