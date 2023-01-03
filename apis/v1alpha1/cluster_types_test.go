/*
Copyright 2023 The Cockroach Authors

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

package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/client/clientset/versioned"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestCrdbCluster(t *testing.T) {
	env := env.NewEnv(runtime.NewSchemeBuilder(AddToScheme))

	env.Start()
	defer env.Stop()

	key := types.NamespacedName{
		Name:      "foo",
		Namespace: "default",
	}

	given := &CrdbCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: CrdbClusterSpec{Nodes: 3},
	}

	ctx := context.TODO()
	client := versioned.NewForConfigOrDie(env.Config).CrdbV1alpha1().CrdbClusters(key.Namespace)

	// create a new cluster
	created, err := client.Create(ctx, given, metav1.CreateOptions{})
	require.NoError(t, err)

	// look it up
	found, err := client.Get(ctx, key.Name, metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, created, found)

	// delete it
	require.NoError(t, client.Delete(ctx, key.Name, metav1.DeleteOptions{}))

	// ensure it's gone
	found, err = client.Get(ctx, key.Name, metav1.GetOptions{})
	require.True(t, errors.IsNotFound(err))
	require.Empty(t, found)
}
