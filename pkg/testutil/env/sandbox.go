package env

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/yaml"
	"sort"
)

const (
	DefaultNsName = "crdb-test-"
)

func NewSandboxOrDie(env *ActiveEnv) Sandbox {
	s := Sandbox{
		env: env,

		Namespace: DefaultNsName + rand.String(6),
	}

	if err := createNamespace(s); err != nil {
		panic(err)
	}

	return s
}

type Sandbox struct {
	env *ActiveEnv

	Namespace string
}

type namespaceCreatable interface {
	RuntimeObject(string) apiruntime.Object
}

func (s Sandbox) Create(nc namespaceCreatable) error {
	return s.env.Client.Create(context.TODO(), nc.RuntimeObject(s.Namespace))
}


func (s Sandbox) Cleanup() {
	dp := metav1.DeletePropagationForeground
	opts := &metav1.DeleteOptions{PropagationPolicy: &dp}
	nss := s.env.Clientset.CoreV1().Namespaces()
	if err := nss.Delete(s.Namespace, opts); err != nil {
		fmt.Println("failed to cleanup namespace", s.Namespace)
	}
}

func createNamespace(s Sandbox) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Namespace,
		},
	}

	if _, err := s.env.Clientset.CoreV1().Namespaces().Create(ns); err != nil {
		return errors.Wrapf(err, "failed to create namespace: %s", s.Namespace)
	}

	return nil
}


func NewDiffingSandboxOrDie(env *ActiveEnv) DiffingSandbox {
	s := NewSandboxOrDie(env)

	return DiffingSandbox{
		Sandbox:  s,
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

	return oo.Diff(ds.originalObjs), nil
}


func listAllObjs(s Sandbox) (objList, error) {
	var l objList
	for _, gvr := range s.env.resources {
		res := s.env.namespaceableResource(gvr)

		list, err := res.Namespace(s.Namespace).List(metav1.ListOptions{})
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
		filterUnnecessary(&u)

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
		if agvk.Group == v1alpha1.GroupVersion.Group {
			return true
		}

		if bgvk.Group == v1alpha1.GroupVersion.Group {
			return false
		}
	}

	if a.GetKind() != b.GetKind() {
		return a.GetKind() < b.GetKind()
	}

	return a.GetName() < b.GetName()
}

func filterUnnecessary(u *unstructured.Unstructured) {
	unstructured.RemoveNestedField(u.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(u.Object, "metadata", "generation")
	unstructured.RemoveNestedField(u.Object, "metadata", "namespace")
	unstructured.RemoveNestedField(u.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(u.Object, "metadata", "selfLink")
	unstructured.RemoveNestedField(u.Object, "metadata", "uid")
}