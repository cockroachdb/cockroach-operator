package actor_test

import (
	"context"
	"github.com/cockroachlabs/crdb-operator/pkg/actor"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"testing"
)

type key struct {
	Resource, Name string
}

type callTracker map[key]int

func (t callTracker) calledOnceFor(resource, name string) error {
	k := key{Resource: resource, Name: name}
	t[k]++

	if t[k] == 2 {
		return errors.Errorf("was called more than once for %s", name)
	}

	return nil
}

func TestDeploysNotInitalizedCluster(t *testing.T) {
	actor.Log = zapr.NewLogger(zaptest.NewLogger(t)).WithName("deploy-test")

	var expected, actual callTracker = make(map[key]int), make(map[key]int)
	_ = expected.calledOnceFor("services", "default/cockroachdb")
	_ = expected.calledOnceFor("services", "default/cockroachdb-public")
	_ = expected.calledOnceFor("statefulsets", "default/cockroachdb")

	scheme := testutil.InitScheme(t)

	client := testutil.NewFakeClient(scheme)
	client.AddReactor("create", "*",
		func(action testutil.Action) (bool, error) {
			if err := actual.calledOnceFor(action.GVR().Resource, action.Key().String()); err != nil {
				return true, err
			}

			return false, nil
		})

	cluster := testutil.NewBuilder("cockroachdb").
		Namespaced("default").
		WithEmptyDirDataStore().
		WithNodeCount(1).Cluster()

	deploy := actor.NewDeploy(scheme, client)
	require.True(t, deploy.Handles(cluster.Status().Conditions))

	// 3 is the number of resources we expect to be created. The action should be repeated as it is
	// restarted on successful creation or update
	for i := 0; i < 3; i++ {
		assert.NoError(t, deploy.Act(actor.ContextWithCancelFn(context.TODO(), func() {}), cluster))
	}

	assert.Equal(t, expected, actual)
}
