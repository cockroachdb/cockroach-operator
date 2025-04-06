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
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/yaml"
)

const (
	DefaultNsName = "crdb-test-"
)

func NewSandbox(t *testing.T, env *ActiveEnv) Sandbox {
	ns := DefaultNsName + rand.String(6)
	td := time.Duration(0)

	mgr, err := ctrl.NewManager(env.k8s.Cfg, ctrl.Options{
		Scheme:                  env.scheme,
		Namespace:               ns,
		MetricsBindAddress:      "0", // disable metrics serving
		GracefulShutdownTimeout: &td,
	})
	require.NoError(t, err)

	s := Sandbox{
		env:       env,
		Namespace: ns,
		Mgr:       mgr,
	}

	require.NoError(t, createNamespace(s))
	t.Cleanup(s.Cleanup)

	return s
}

type Sandbox struct {
	env *ActiveEnv
	Mgr ctrl.Manager

	Namespace string
}

func (s Sandbox) setNamespaceIfMissing(obj client.Object) client.Object {
	accessor, _ := meta.Accessor(obj)
	if accessor.GetNamespace() == "" {
		runtimeObject := obj.DeepCopyObject()
		accessor, _ = meta.Accessor(runtimeObject)
		accessor.SetNamespace(s.Namespace)
		obj, _ := runtimeObject.(client.Object)
		return obj
	}

	return obj
}

func (s Sandbox) Create(obj client.Object) error {
	obj = s.setNamespaceIfMissing(obj)

	return s.env.Create(context.TODO(), obj)
}

func (s Sandbox) Update(obj client.Object) error {
	obj = s.setNamespaceIfMissing(obj)

	return s.env.Update(context.TODO(), obj)
}

func (s Sandbox) Patch(obj client.Object, patch client.Patch) error {
	obj = s.setNamespaceIfMissing(obj)

	return s.env.Patch(context.TODO(), obj, patch)
}

func (s Sandbox) Delete(obj client.Object) error {
	obj = s.setNamespaceIfMissing(obj)

	return s.env.Delete(context.TODO(), obj)
}

func (s Sandbox) Get(o client.Object) error {
	accessor, err := meta.Accessor(o)
	if err != nil {
		return err
	}

	key := types.NamespacedName{
		Namespace: s.Namespace,
		Name:      accessor.GetName(),
	}

	return s.env.Get(context.TODO(), key, o)
}

func (s Sandbox) List(list client.ObjectList, labels map[string]string) error {
	ns := client.InNamespace(s.Namespace)
	matchingLabels := client.MatchingLabels(labels)

	return s.env.List(context.TODO(), list, ns, matchingLabels)
}

func (s Sandbox) Cleanup() {
	dp := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &dp}
	nss := s.env.Clientset.CoreV1().Namespaces()
	if err := nss.Delete(context.TODO(), s.Namespace, opts); err != nil {
		fmt.Println("failed to cleanup namespace", s.Namespace)
	}
}

func (s Sandbox) StartManager(t *testing.T, maker func(ctrl.Manager) error) {
	require.NoError(t, maker(s.Mgr))
	t.Cleanup(startCtrlMgr(t, s.Mgr))
}

func createNamespace(s Sandbox) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Namespace,
		},
	}

	if _, err := s.env.Clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{}); err != nil {
		return errors.Wrapf(err, "failed to create namespace: %s", s.Namespace)
	}

	return nil
}

// nolint
func startCtrlMgr(t *testing.T, mgr manager.Manager) func() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		require.NoError(t, mgr.Start(ctx))
	}()

	return cancel
}

func NewDiffingSandbox(t *testing.T, env *ActiveEnv) DiffingSandbox {
	s := NewSandbox(t, env)

	return DiffingSandbox{
		Sandbox:      s,
		originalObjs: listAllObjsOrDie(s),
	}
}

type DiffingSandbox struct {
	Sandbox

	originalObjs objList
}

func (ds *DiffingSandbox) Diff() (string, error) {
	oo, err := listAllObjs(ds.Sandbox)
	if err != nil {
		return "", err
	}

	diff := oo.Diff(ds.originalObjs)

	redacted := strings.ReplaceAll(diff, ds.Namespace, "[sandbox_namespace]")

	return redacted, nil
}

func listAllObjs(s Sandbox) (objList, error) {
	var l objList
	for _, gvr := range s.env.resources {
		res := s.env.namespaceableResource(gvr)

		list, err := res.Namespace(s.Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list objects in namespace %s", s.Namespace)
		}

		l = append(l, list.Items...)
	}

	sort.Sort(l)

	return l, nil
}

func listAllObjsOrDie(s Sandbox) objList {
	oo, err := listAllObjs(s)
	if err != nil {
		panic(err)
	}

	return oo
}

type objList []unstructured.Unstructured

func (l objList) Diff(other objList) string {
	diff := objList{}

OUTER:
	for _, o1 := range l {
		for _, o2 := range other {
			if o1.GroupVersionKind() == o2.GroupVersionKind() && o1.GetName() == o2.GetName() {
				continue OUTER
			}
		}

		diff = append(diff, o1)
	}

	return string(diff.ToYamlOrDie())
}

func (l objList) ToYamlOrDie() []byte {
	var out bytes.Buffer
	for _, u := range l {
		if ignoreObject(&u) {
			continue
		}

		stripUnnecessaryDetails(&u)

		bs, err := yaml.Marshal(u.Object)
		if err != nil {
			panic(err)
		}
		out.WriteString("---\n")
		out.Write(bs)
		out.WriteRune('\n')
	}

	return out.Bytes()
}

func (l objList) Len() int      { return len(l) }
func (l objList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l objList) Less(i, j int) bool {
	a, b := &l[i], &l[j]

	if a.GetNamespace() != b.GetNamespace() {
		return a.GetNamespace() < b.GetNamespace()
	}

	agvk, bgvk := a.GroupVersionKind(), b.GroupVersionKind()

	if agvk.Group != bgvk.Group {
		if agvk.Group == api.SchemeGroupVersion.Group {
			return true
		}

		if bgvk.Group == api.SchemeGroupVersion.Group {
			return false
		}
	}

	if a.GetKind() != b.GetKind() {
		return a.GetKind() < b.GetKind()
	}

	return a.GetName() < b.GetName()
}

func ignoreObject(u *unstructured.Unstructured) bool {
	// Default account secret && cockroachdb-sa secret
	return u.GetKind() == "Secret" && strings.Contains(u.GetName(), "-token-")
}

func stripUnnecessaryDetails(u *unstructured.Unstructured) {
	if u.GetKind() == "Pod" && u.GetAPIVersion() == "v1" {
		unstructured.RemoveNestedField(u.Object, "spec", "nodeName")
		unstructured.RemoveNestedField(u.Object, "spec", "hostname")

		replaceDefaultTokenNames(u)
		filterPodLabels(u)
	}

	if u.GetKind() == "Service" && u.GetAPIVersion() == "v1" {
		replaceServiceIP(u)
	}

	if u.GetKind() == "Secret" && u.GetAPIVersion() == "v1" {
		replaceSecretContent(u)
	}

	aa := u.GetAnnotations()
	if aa != nil {
		delete(aa, kube.LastAppliedAnnotation)
		if len(aa) > 0 {
			u.SetAnnotations(aa)
		} else {
			unstructured.RemoveNestedField(u.Object, "metadata", "annotations")
		}
	}

	unstructured.RemoveNestedField(u.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(u.Object, "metadata", "generation")
	unstructured.RemoveNestedField(u.Object, "metadata", "namespace")
	unstructured.RemoveNestedField(u.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(u.Object, "metadata", "selfLink")
	unstructured.RemoveNestedField(u.Object, "metadata", "uid")
	unstructured.RemoveNestedField(u.Object, "metadata", "generateName")
	unstructured.RemoveNestedField(u.Object, "metadata", "ownerReferences")
	unstructured.RemoveNestedField(u.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(u.Object, "status")
}

func replaceDefaultTokenNames(u *unstructured.Unstructured) {
	containers, ok, err := unstructured.NestedSlice(u.Object, "spec", "containers")
	if (err != nil) || !ok {
		return
	}

	replacements := make(map[string]string)

	var newContainers []interface{}
	for i, rawContainer := range containers {
		container, _ := rawContainer.(map[string]interface{})

		volumeMounts, ok, _ := unstructured.NestedSlice(container, "volumeMounts")
		if ok {
			var newVolumeMounts []interface{}
			for _, rawVolumeMount := range volumeMounts {
				volumeMount, _ := rawVolumeMount.(map[string]interface{})

				vmName, _ := volumeMount["name"].(string)

				if strings.HasPrefix(vmName, "default-token-") {
					newName := fmt.Sprintf("default-token-%d", i)
					replacements[vmName] = newName
					volumeMount["name"] = newName
				}

				newVolumeMounts = append(newVolumeMounts, volumeMount)
			}

			_ = unstructured.SetNestedSlice(container, newVolumeMounts, "volumeMounts")
		}

		newContainers = append(newContainers, container)
	}

	_ = unstructured.SetNestedSlice(u.Object, newContainers, "spec", "containers")

	volumes, ok, err := unstructured.NestedSlice(u.Object, "spec", "volumes")
	var newVolumes []interface{}
	if (err == nil) && ok {
		for _, rawVolume := range volumes {
			volume, mapOk := rawVolume.(map[string]interface{})
			name, nameOk := volume["name"].(string)
			secretMap, secretOk, _ := unstructured.NestedMap(volume, "secret")
			if mapOk && nameOk && secretOk {
				_, replaced := replacements[name]
				_, hasSecretName := secretMap["secretName"]
				if replaced && hasSecretName {
					volume["name"] = replacements[name]
					secretMap["secretName"] = replacements[name]

					_ = unstructured.SetNestedMap(volume, secretMap, "secret")
				}
			}

			newVolumes = append(newVolumes, volume)
		}
	}

	_ = unstructured.SetNestedSlice(u.Object, newVolumes, "spec", "volumes")
}

func filterPodLabels(u *unstructured.Unstructured) {
	labels := u.GetLabels()
	delete(labels, "controller-revision-hash")
	u.SetLabels(labels)
}

func replaceServiceIP(u *unstructured.Unstructured) {
	clusterIP, ok, err := unstructured.NestedString(u.Object, "spec", "clusterIP")
	if (err != nil) || !ok {
		return
	}

	if clusterIP != "None" {
		_ = unstructured.SetNestedField(u.Object, "[some_ip]", "spec", "clusterIP")
	}
}

func replaceSecretContent(u *unstructured.Unstructured) {
	data, ok, err := unstructured.NestedStringMap(u.Object, "data")
	if (err != nil) || !ok {
		return
	}

	for k := range data {
		data[k] = "[replaced]"
	}

	_ = unstructured.SetNestedStringMap(u.Object, data, "data")
}
