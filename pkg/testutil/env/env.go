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

package env

import (
	"flag"
	"fmt"
	"os"
	"time"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	customClient "github.com/cockroachdb/cockroach-operator/pkg/client/clientset/versioned"
	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // GCP auth support
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// testBinaries is used to set KUBEBUILDER_ASSETS variable to ensure things like etcd are available to all test
// environments.
var testBinaries = flag.String("binaries", "hack/bin", "")

func NewEnv(builder apiruntime.SchemeBuilder) *Env {
	flag.Parse()

	// ensure hack/bin is added to the path and KUBEBUILDER_ASSETS
	p := ExpandPath(*testBinaries)
	os.Setenv("KUBEBUILDER_ASSETS", p)
	PrependToPath(p)

	scheme := apiruntime.NewScheme()

	if err := kscheme.AddToScheme(scheme); err != nil {
		panic(err)
	}

	if err := builder.AddToScheme(scheme); err != nil {
		panic(err)
	}

	t := envtest.Environment{
		CRDDirectoryPaths: []string{
			ExpandPath("config", "crd"),
			ExpandPath("config", "webhook"),
		},
		CRDInstallOptions: envtest.CRDInstallOptions{
			CleanUpAfterUse: true,
		},
	}

	return &Env{
		Environment: t,
		Scheme:      scheme,
	}
}

type Env struct {
	envtest.Environment
	Scheme *apiruntime.Scheme
}

func (env *Env) Start() *ActiveEnv {
	if _, err := env.Environment.Start(); err != nil {
		panic(err)
	}

	dc, err := dynamic.NewForConfig(env.Environment.Config)
	if err != nil {
		panic(err)
	}

	c, err := client.New(env.Environment.Config, client.Options{Scheme: env.Scheme})
	if err != nil {
		panic(err)
	}

	k8s := &k8s{
		Client:    c,
		Clientset: kubernetes.NewForConfigOrDie(env.Environment.Config),
		Interface: dc,
		Cfg:       env.Environment.Config,

		CustomClient: customClient.NewForConfigOrDie(env.Environment.Config),
	}

	resources, err := loadResources(k8s)
	if err != nil {
		panic(err)
	}

	return &ActiveEnv{
		k8s:       k8s,
		scheme:    env.Scheme,
		resources: resources,
	}
}

type k8s struct {
	client.Client
	*kubernetes.Clientset
	dynamic.Interface
	Cfg          *rest.Config
	CustomClient *customClient.Clientset
}

func (k k8s) namespaceableResource(gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	return k.Interface.Resource(gvr)
}

type ActiveEnv struct {
	*k8s
	scheme    *apiruntime.Scheme
	resources []schema.GroupVersionResource
}

func CreateActiveEnvForTest() *Env {
	os.Setenv("USE_EXISTING_CLUSTER", "true")
	return NewEnv(runtime.NewSchemeBuilder(api.AddToScheme))
}

func loadResources(k *k8s) ([]schema.GroupVersionResource, error) {
	lists, err := k.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch preferred server resource")
	}

	var resources []schema.GroupVersionResource
	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}

		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}

		for _, r := range list.APIResources {
			if len(r.Verbs) == 0 {
				continue
			}

			if len(r.Verbs) > 0 && !in(r.Verbs, "list") {
				continue
			}

			if !r.Namespaced {
				continue
			}

			if filteredKind(r.Kind) {
				continue
			}

			resources = append(resources, gv.WithResource(r.Name))
		}
	}

	return resources, nil
}

func (env *Env) Stop() {
	defer func() {
		if err := env.Environment.Stop(); err != nil {
			fmt.Println(err)
		}
	}()

	if err := envtest.UninstallCRDs(env.Environment.Config, env.Environment.CRDInstallOptions); err != nil {
		panic(err)
	}

	// Wait for CRDs to be uninstalled
	time.Sleep(10 * time.Second)
}

func (env *Env) StopAndExit(code int) {
	env.Stop()
	os.Exit(code)
}

func in(haystack []string, needle string) bool {
	for _, s := range haystack {
		if needle == s {
			return true
		}
	}
	return false
}

func filteredKind(k string) bool {
	return k == "PodMetrics" || k == "Event" || k == "Endpoints" ||
		k == "ControllerRevision" || k == "ServiceAccount" || k == "EndpointSlice"
}
