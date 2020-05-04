package e2e

import (
	"flag"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/actor"
	"github.com/cockroachlabs/crdb-operator/pkg/controller"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	testenv "github.com/cockroachlabs/crdb-operator/pkg/testutil/env"
	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
	"time"
)

var updateOpt = flag.Bool("update", false, "update the golden files of this test")

var env *testenv.ActiveEnv

type Step struct {
	name string
	test func(t *testing.T)
}

type Steps []Step

func (ss Steps) WithStep(s Step) Steps {
	return append(ss, s)
}

func (ss Steps) Run(t *testing.T) {
	for _, s := range ss {
		if !t.Run(s.name, s.test) {
			t.FailNow()
		}
	}
}

func TestMain(m *testing.M) {
	flag.Parse()

	e := testenv.NewEnv(runtime.NewSchemeBuilder(api.AddToScheme),
		filepath.Join("..", "config", "crd", "bases"))

	env = e.Start()

	e.StopAndExit(m.Run())
}

func TestCreatesInsecureCluster(t *testing.T) {
	testLog := zapr.NewLogger(zaptest.NewLogger(t))

	actor.Log = testLog

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	sb := testenv.NewDiffingSandbox(t, env)
	sb.StartManager(t, controller.InitClusterReconcilerWithLogger(testLog))

	b := testutil.NewBuilder("crdb").WithNodeCount(1).WithEmptyDirDataStore()

	create := Step{
		name: "creates 1-node cluster",
		test: func(t *testing.T) {
			require.NoError(t, sb.Create(b))

			require.NoError(t, wait.PollImmediate(10*time.Second, 180*time.Second, func() (bool, error) {
				cluster := b.Cluster()

				expectedConditions := []api.ClusterCondition{
					{
						Type:   api.InitializedCondition,
						Status: metav1.ConditionTrue,
					},
				}

				actual := resource.ClusterPlaceholder(cluster.Name())
				if err := sb.Get(actual); err != nil {
					return false, err
				}

				actualConditions := actual.Status.DeepCopy().Conditions
				var emptyTime metav1.Time
				for i := range actualConditions {
					actualConditions[i].LastTransitionTime = emptyTime
				}

				if !cmp.Equal(expectedConditions, actualConditions) {
					t.Logf("expected condtitions do not match. expected: %+v, actual: %+v",
						expectedConditions, actualConditions)
					return false, nil
				}

				ss := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: cluster.StatefulSetName(),
					},
				}

				if err := sb.Get(ss); err != nil {
					if apierrors.IsNotFound(err) {
						t.Logf("stateful set is not found")
						return false, nil
					}
					return false, client.IgnoreNotFound(err)
				}

				return ss.Status.ReadyReplicas == ss.Status.Replicas, nil
			}))

			state, err := sb.Diff()
			require.NoError(t, err)

			expected := testutil.ReadOrUpdateGoldenFile(t, state, *updateOpt)

			testutil.AssertDiff(t, expected, state)
		},
	}

	steps := Steps{create}

	steps.Run(t)
}
