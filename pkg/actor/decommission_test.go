/*
Copyright 2021 The Cockroach Authors

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

package actor

import (
	"testing"

	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecommisionFeatureFlag(t *testing.T) {
	// FeatureGate is currently enabled as of 1.7.13 GA release
	assert.True(t, utilfeature.DefaultMutableFeatureGate.Enabled(features.Decommission), "deco is enabled for GA")
	// FeatureGate is currently disabled and is in alpha
	assert.False(t, utilfeature.DefaultMutableFeatureGate.Enabled(features.AutoPrunePVC), "AutoPrunePVc is disabled for GA")

	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithUID("cockroachdb-uid").
		WithPVDataStore("1Gi", "standard" /* default storage class in KIND */).
		WithNodeCount(1).Cluster()
	cluster.SetTrue(api.InitializedCondition)

	scheme := testutil.InitScheme(t)
	client := testutil.NewFakeClient(scheme)
	deco := newDecommission(scheme, client, nil)

	require.True(t, deco.Handles(cluster.Status().Conditions))

}
