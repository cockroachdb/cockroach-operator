package controller_test

import (
	"context"
	"errors"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/actor"
	"github.com/cockroachlabs/crdb-operator/pkg/controller"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

type fakeActor struct {
	handlesCalled bool
	cancelCtx     bool
	err           error
}

func (a *fakeActor) Handles(_ []api.ClusterCondition) bool {
	a.handlesCalled = true
	return true
}

func (a *fakeActor) Act(ctx context.Context, _ *resource.Cluster) error {
	if a.cancelCtx {
		actor.CancelLoop(ctx)
	}
	return a.err
}

func TestReconcile(t *testing.T) {
	scheme := testutil.InitScheme(t)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	cluster := testutil.NewBuilder("cluster").Namespaced(ns.Name).WithNodeCount(1).Cr()

	objs := []runtime.Object{
		ns,
		cluster,
	}

	cl := fake.NewFakeClientWithScheme(scheme, objs...)
	log := zapr.NewLogger(zaptest.NewLogger(t)).WithName("cluster-controller-test")
	req := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}}

	tests := []struct {
		name    string
		action  fakeActor
		want    ctrl.Result
		wantErr string
	}{
		{
			name: "reconcile action fails",
			action: fakeActor{
				err: errors.New("failed to reconcile resource"),
			},
			want:    ctrl.Result{Requeue: false},
			wantErr: "failed to reconcile resource",
		},
		{
			name:    "reconcile action updates owned resource successfully",
			action:  fakeActor{},
			want:    ctrl.Result{Requeue: false},
			wantErr: "",
		},
		{
			name: "reconcile action cancels the context",
			action: fakeActor{
				cancelCtx: true,
			},
			want:    ctrl.Result{Requeue: false},
			wantErr: "",
		},
		{
			name: "reconcile action fails to probe expected condition",
			action: fakeActor{
				err: actor.NotReadyErr{Err: errors.New("not ready")},
			},
			want:    ctrl.Result{RequeueAfter: 5 * time.Second},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &controller.ClusterReconciler{
				Client: cl,
				Log:    log,
				Scheme: scheme,
				Actions: []actor.Actor{
					&tt.action,
				},
			}

			actual, err := r.Reconcile(req)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			}

			assert.Equal(t, tt.want, actual)
		})
	}

}
