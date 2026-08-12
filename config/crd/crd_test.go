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

package crd_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

const crdbClusterDeprecationWarning = "crdb.cockroachlabs.com/v1alpha1 CrdbCluster is deprecated; see https://www.cockroachlabs.com/docs/v26.2/kubernetes-deprecation-notice"

func TestCrdbClusterV1Alpha1IsDeprecated(t *testing.T) {
	data, err := os.ReadFile("bases/crdb.cockroachlabs.com_crdbclusters.yaml")
	require.NoError(t, err)

	var crd apiextensionsv1.CustomResourceDefinition
	require.NoError(t, yaml.Unmarshal(data, &crd))

	for _, version := range crd.Spec.Versions {
		if version.Name != "v1alpha1" {
			continue
		}

		require.True(t, version.Deprecated)
		require.NotNil(t, version.DeprecationWarning)
		require.Equal(t, crdbClusterDeprecationWarning, *version.DeprecationWarning)
		return
	}

	t.Fatal("v1alpha1 CRD version not found")
}
