/*
Copyright 2022 The Cockroach Authors

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

package scale

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakeclient "k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestStatefulSetIsRunning(t *testing.T) {
	stsReplicas := int32(3)
	cltSet := fakeclient.NewSimpleClientset()

	// Test error if no statefulset exists
	require.Errorf(t, StatefulSetIsRunning(context.TODO(), cltSet, "crdb", "crdb-sts"),
		"failed to get statefulset: %s", "crdb-sts")

	sts := statefulSet(stsReplicas)

	require.NoError(t, cltSet.Tracker().Add(&sts))

	cltSet.PrependReactor("*", "*", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		tracker := cltSet.Tracker()
		gvr := action.GetResource()
		ns := action.GetNamespace()
		verb := action.GetVerb()

		switch verb {
		case "get":
			getAction := action.(clienttesting.GetAction)
			obj, err := tracker.Get(gvr, ns, getAction.GetName())
			require.NoError(t, err)

			return true, obj, nil
		case "list":
			obj, err := tracker.List(gvr, schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
				ns) // Tracker ignores labels passed in and only checks the namespace
			require.NoError(t, err)

			return true, obj, nil
		default:
			t.Logf("Unexpected action %v", action)
			t.Fail()
		}

		return false, nil, nil
	})

	// No replicas running
	require.Errorf(t, StatefulSetIsRunning(context.TODO(), cltSet, "crdb", "crdb-sts"),
		"statefulset replicas not yet reconciled. have %d expected %d",
		sts.Status.Replicas,
		sts.Spec.Replicas)

	// Set Status Replicas Ready to be number of replicas defined in the spec
	sts.Status.Replicas = stsReplicas
	require.NoError(t, cltSet.Tracker().Update(sts.GroupVersionKind().GroupVersion().WithResource("statefulsets"), &sts, "crdb"))

	require.NoError(t, addPodsToStatefulSet(stsReplicas, sts, cltSet))

	require.NoError(t, StatefulSetIsRunning(context.TODO(), cltSet, "crdb", "crdb-sts"))
}

func TestIsStatefulSetReadyToServe(t *testing.T) {
	stsReplicas := int32(3)

	sts := statefulSet(stsReplicas)

	cltSet := fakeclient.NewSimpleClientset()
	require.NoError(t, cltSet.Tracker().Add(&sts))

	cltSet.PrependReactor("*", "*", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		tracker := cltSet.Tracker()
		gvr := action.GetResource()
		ns := action.GetNamespace()
		verb := action.GetVerb()

		switch verb {
		case "get":
			getAction := action.(clienttesting.GetAction)
			obj, err := tracker.Get(gvr, ns, getAction.GetName())
			require.NoError(t, err)

			return true, obj, nil
		default:
			t.Logf("Unexpected action %v", action)
			t.Fail()
		}

		return false, nil, nil
	})

	// No replicas running
	require.Error(t, IsStatefulSetReadyToServe(context.TODO(), cltSet, "crdb", "crdb-sts", stsReplicas))

	// Set Status Replicas Ready to be number of replicas defined in the spec
	sts.Status.Replicas = stsReplicas
	sts.Status.ReadyReplicas = stsReplicas
	require.NoError(t, cltSet.Tracker().Update(sts.GroupVersionKind().GroupVersion().WithResource("statefulsets"), &sts, "crdb"))
	require.NoError(t, IsStatefulSetReadyToServe(context.TODO(), cltSet, "crdb", "crdb-sts", stsReplicas))

	// Err if there are more replicas than in the spec
	sts.Status.Replicas = stsReplicas * 2
	require.NoError(t, cltSet.Tracker().Update(sts.GroupVersionKind().GroupVersion().WithResource("statefulsets"), &sts, "crdb"))
	require.Error(t, IsStatefulSetReadyToServe(context.TODO(), cltSet, "crdb", "crdb-sts", stsReplicas))

}

func statefulSet(stsReplicas int32) appsv1.StatefulSet {
	return appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "statefulset",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string),
			Name:        "crdb-sts",
			Namespace:   "crdb",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/component": "crdb-sts",
					"app.kubernetes.io/instance":  "crdb-sts",
					"app.kubernetes.io/name":      "crdb-sts",
				},
				MatchExpressions: nil,
			},
			Replicas: &stsReplicas,
		},
	}
}

func addPodsToStatefulSet(stsReplicas int32, sts appsv1.StatefulSet, cltSet *fakeclient.Clientset) error {
	// Create some pods to look up
	var i int32
	for i = 0; i < stsReplicas; i++ {
		pod := v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%v", sts.Name, i),
				Namespace: sts.Namespace,
				Labels:    sts.Spec.Selector.MatchLabels,
			},
			Status: v1.PodStatus{
				Phase: "Running",
			},
		}

		if err := cltSet.Tracker().Add(&pod); err != nil {
			return err
		}
	}
	return nil
}
