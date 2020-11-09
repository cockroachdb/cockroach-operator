package scale

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8s_testing "k8s.io/client-go/testing"
)

func TestPersistentVolumePruner_Prune(t *testing.T) {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cockroachdb",
			Namespace: "testns",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "cockroach",
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "datadir",
						Namespace: "testns",
					},
				},
			},
		},
	}

	testCases := []struct {
		Name        string
		Replicas    int32
		PVCs        int32
		Setup       func(*fake.Clientset)
		HandleError func(*testing.T, *fake.Clientset, error)
	}{
		{
			Name:     "No extra PVCs",
			Replicas: 5,
			PVCs:     5,
		},
		{
			Name:     "A couple extra PVCs",
			Replicas: 5,
			PVCs:     7,
		},
		{
			Name:     "Concurrent modification no replica change",
			Replicas: 5,
			PVCs:     7,
			Setup: func(cs *fake.Clientset) {
				cs.Fake.PrependWatchReactor("*", func(action k8s_testing.Action) (bool, watch.Interface, error) {
					w := watch.NewRaceFreeFake()

					sts := sts.DeepCopy()
					replicas := int32(5)
					sts.Spec.Replicas = &replicas

					w.Modify(sts)

					return true, w, nil
				})
			},
		},
		{
			Name:     "Concurrent modification",
			Replicas: 5,
			PVCs:     7,
			Setup: func(cs *fake.Clientset) {
				cs.Fake.PrependWatchReactor("*", func(action k8s_testing.Action) (bool, watch.Interface, error) {
					w := watch.NewRaceFreeFake()

					sts := sts.DeepCopy()
					replicas := int32(6)
					sts.Spec.Replicas = &replicas

					w.Modify(sts)

					return true, w, nil
				})
			},
			HandleError: func(t *testing.T, cs *fake.Clientset, err error) {
				require.EqualError(t, err, "concurrent statefulset modification detected")
			},
		},
		{
			Name:     "unexpected events",
			Replicas: 5,
			PVCs:     7,
			Setup: func(cs *fake.Clientset) {
				cs.Fake.PrependWatchReactor("*", func(action k8s_testing.Action) (bool, watch.Interface, error) {
					w := watch.NewRaceFreeFake()

					w.Delete(sts.DeepCopy())

					return true, w, nil
				})
			},
			HandleError: func(t *testing.T, cs *fake.Clientset, err error) {
				require.EqualError(t, err, "concurrent statefulset modification detected")
			},
		},
		{
			Name:     "unexpected PVC",
			Replicas: 5,
			PVCs:     7,
			Setup: func(cs *fake.Clientset) {
				_ = cs.Tracker().Add(&corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						// An _unexpected name_.
						Name:      "the-spanish-inquisition",
						Namespace: "testns",
						Labels: map[string]string{
							"app": "cockroach",
						},
					},
				})
			},
			HandleError: func(t *testing.T, cs *fake.Clientset, err error) {
				// The operation was still successful
				require.NoError(t, err)

				// But we have 5 + 1 (the unexpected pvc) left over
				pvcs, err := cs.CoreV1().PersistentVolumeClaims("testns").List(metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, pvcs.Items, 6)

				found := false
				for _, pvc := range pvcs.Items {
					if pvc.Name == "the-spanish-inquisition" {
						found = true
						break
					}
				}
				require.True(t, found)
			},
		},
		{
			Name:     "long time to delete",
			Replicas: 5,
			PVCs:     7,
			Setup: func(cs *fake.Clientset) {
				cs.PrependReactor("delete", "*", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
					// Delay deletion for 1 second. This is long enough to
					// fail our tests if we're not waiting for deletion.
					// Technically, we should set the state to terminating but
					// our code doesn't actually check the state.
					go func() {
						time.Sleep(time.Second)
						deleteAction := action.(k8s_testing.DeleteActionImpl)

						if err := cs.Tracker().Delete(
							deleteAction.GetResource(),
							deleteAction.Namespace,
							deleteAction.Name,
						); err != nil {
							panic(err)
						}
					}()

					// Pretend that we've acknowledged the deletion
					return true, nil, nil
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			logger := zaptest.NewLogger(t)

			sts := sts.DeepCopy()
			sts.Spec.Replicas = &tc.Replicas

			objects := []runtime.Object{sts}
			for i := int32(0); i < tc.PVCs; i++ {
				objects = append(objects, &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("datadir-cockroachdb-%d", i),
						Namespace: "testns",
						Labels: map[string]string{
							"app": "cockroach",
						},
					},
				})
			}

			cs := fake.NewSimpleClientset(objects...)

			// Nasty hack, to add some latency into k8s calls. In a real world
			// scenario we would there wouldn't be much of a worry here. In
			// tests, the fake reactor can move a bit faster than the watch
			// events which leads to flakey tests. A single millisecond appears
			// to be long enough to ensure success.
			cs.PrependReactor("*", "*", func(action k8s_testing.Action) (bool, runtime.Object, error) {
				time.Sleep(1 * time.Millisecond)
				return false, nil, nil
			})

			if tc.Setup != nil {
				tc.Setup(cs)
			}

			pruner := PersistentVolumePruner{
				Namespace:   "testns",
				StatefulSet: "cockroachdb",
				ClientSet:   cs,
				Logger:      logger,
			}

			err := pruner.Prune(ctx)
			if tc.HandleError != nil {
				tc.HandleError(t, cs, err)
				return
			}

			require.NoError(t, err)

			pvcs, err := cs.CoreV1().PersistentVolumeClaims("testns").List(ctx, metav1.ListOptions{})
			require.NoError(t, err)
			require.Len(t, pvcs.Items, int(tc.Replicas))
		})
	}
}
