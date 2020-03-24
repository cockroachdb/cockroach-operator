package controller_test

import (
	crdbv1alpha1 "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var sc = runtime.NewScheme()

func init() {
	if err := corev1.AddToScheme(sc); err != nil {
		panic(err)
	}

	if err := crdbv1alpha1.AddToScheme(sc); err != nil {
		panic(err)
	}

	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(os.Stdout)))
}

func TestReconcile(t *testing.T) {
	cluster := &crdbv1alpha1.CrdbCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster1",
			Namespace: "test-namespace",
		},
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	objs := []runtime.Object{
		ns,
		cluster,
	}

	cl := fake.NewFakeClientWithScheme(sc, objs...)

	r := &controller.CrdbClusterReconciler{Client: cl, Log: ctrl.Log.WithName("test"), Scheme: sc}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "cluster1",
			Namespace: "test-namespace",
		},
	}

	_, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
}
