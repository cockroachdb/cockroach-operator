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

package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/database"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/testutil"
	testenv "github.com/cockroachdb/cockroach-operator/pkg/testutil/env"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func requireClusterToBeReadyEventually(t *testing.T, sb testenv.DiffingSandbox, b testutil.ClusterBuilder) {
	cluster := b.Cluster()

	err := wait.Poll(10*time.Second, 600*time.Second, func() (bool, error) {
		if initialized, err := clusterIsInitialized(t, sb, cluster.Name()); err != nil || !initialized {
			return false, err
		}

		ss, err := fetchStatefulSet(sb, cluster.StatefulSetName())
		if err != nil {
			return false, err
		}

		if ss == nil {
			t.Logf("stateful set is not found")
			return false, nil
		}

		return statefulSetIsReady(ss), nil
	})
	require.NoError(t, err)
}

func requireDbContainersToUseImage(t *testing.T, sb testenv.DiffingSandbox, cr *api.CrdbCluster) {
	err := wait.Poll(10*time.Second, 180*time.Second, func() (bool, error) {
		pods, err := fetchPodsInStatefulSet(sb, labels.Common(cr).Selector())
		if err != nil {
			return false, err
		}

		if len(pods) < int(cr.Spec.Nodes) {
			return false, nil
		}

		res := testPodsWithPredicate(pods, func(p *corev1.Pod) bool {
			c, err := kube.FindContainer(resource.DbContainerName, &p.Spec)
			if err != nil {
				return false
			}

			return c.Image == cr.Spec.Image.Name
		})

		return res, nil
	})

	require.NoError(t, err)
}

func clusterIsInitialized(t *testing.T, sb testenv.DiffingSandbox, name string) (bool, error) {
	expectedConditions := []api.ClusterCondition{
		{
			Type:   api.NotInitializedCondition,
			Status: metav1.ConditionFalse,
		},
	}

	actual := resource.ClusterPlaceholder(name)
	if err := sb.Get(actual); err != nil {
		t.Logf("failed to fetch current cluster status :(")
		return false, err
	}

	actualConditions := actual.Status.DeepCopy().Conditions

	// Reset condition time as it is not significant for the assertion
	var emptyTime metav1.Time
	for i := range actualConditions {
		actualConditions[i].LastTransitionTime = emptyTime
	}

	if !cmp.Equal(expectedConditions, actualConditions) {
		return false, nil
	}

	return true, nil
}

func clusterIsDecommissioned(t *testing.T, sb testenv.DiffingSandbox, name string) (bool, error) {
	expectedConditions := []api.ClusterCondition{
		{
			Type:   api.DecommissionCondition,
			Status: metav1.ConditionTrue,
		},
	}

	actual := resource.ClusterPlaceholder(name)
	if err := sb.Get(actual); err != nil {
		t.Logf("failed to fetch current cluster status :(")
		return false, err
	}

	actualConditions := actual.Status.DeepCopy().Conditions

	// Reset condition time as it is not significant for the assertion
	var emptyTime metav1.Time
	for i := range actualConditions {
		actualConditions[i].LastTransitionTime = emptyTime
	}
	if !cmp.Equal(expectedConditions, actualConditions) {
		return false, nil
	}

	return true, nil
}

func fetchStatefulSet(sb testenv.DiffingSandbox, name string) (*appsv1.StatefulSet, error) {
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := sb.Get(ss); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return ss, nil
}

func fetchPodsInStatefulSet(sb testenv.DiffingSandbox, labels map[string]string) ([]corev1.Pod, error) {
	var pods corev1.PodList

	if err := sb.List(&pods, labels); err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func testPodsWithPredicate(pods []corev1.Pod, pred func(*corev1.Pod) bool) bool {
	for i := range pods {
		if !pred(&pods[i]) {
			return false
		}
	}

	return true
}

func statefulSetIsReady(ss *appsv1.StatefulSet) bool {
	return ss.Status.ReadyReplicas == ss.Status.Replicas
}

func requireDownGradeOptionSet(t *testing.T, sb testenv.DiffingSandbox, b testutil.ClusterBuilder, version string) {
	sb.Mgr.GetConfig()
	podName := fmt.Sprintf("%s-0.%s", b.Cluster().Name(), b.Cluster().Name())
	conn := &database.DBConnection{
		Ctx:    context.TODO(),
		Client: sb.Mgr.GetClient(),
		Port:   b.Cluster().Spec().GRPCPort,
		UseSSL: true,

		RestConfig:   sb.Mgr.GetConfig(),
		ServiceName:  podName,
		Namespace:    sb.Namespace,
		DatabaseName: "system",

		RunningInsideK8s:            false,
		ClientCertificateSecretName: b.Cluster().ClientTLSSecretName(),
		RootCertificateSecretName:   b.Cluster().NodeTLSSecretName(),
	}

	// Create a new database connection for the update.
	db, err := database.NewDbConnection(conn)
	require.NoError(t, err)
	defer db.Close()

	r := db.QueryRowContext(context.TODO(), "SHOW CLUSTER SETTING cluster.preserve_downgrade_option")
	var value string
	if err := r.Scan(&value); err != nil {
		t.Fatal(err)
	}

	if value == "" {
		t.Errorf("downgrade_option is empty and should be set to %s", version)
	}

	if value != value {
		t.Errorf("downgrade_option is not set to %s, but is set to %s", version, value)
	}

}
func requireDecommissionNode(t *testing.T, sb testenv.DiffingSandbox, b testutil.ClusterBuilder) {
	cluster := b.Cluster()

	err := wait.Poll(10*time.Second, 150*time.Second, func() (bool, error) {
		if initialized, err := clusterIsInitialized(t, sb, cluster.Name()); err != nil || !initialized {
			return false, err
		}

		ss, err := fetchStatefulSet(sb, cluster.StatefulSetName())
		if err != nil {
			return false, err
		}

		if ss == nil {
			t.Logf("stateful set is not found")
			return false, nil
		}
		return statefulSetIsReady(ss) && b.Cr().Spec.Nodes == ss.Status.Replicas, nil
	})
	require.NoError(t, err)
}

func requireDatabaseToFunction(t *testing.T, sb testenv.DiffingSandbox, b testutil.ClusterBuilder) {
	sb.Mgr.GetConfig()
	podName := fmt.Sprintf("%s-0.%s", b.Cluster().Name(), b.Cluster().Name())
	conn := &database.DBConnection{
		Ctx:    context.TODO(),
		Client: sb.Mgr.GetClient(),
		Port:   b.Cluster().Spec().GRPCPort,
		UseSSL: true,

		RestConfig:   sb.Mgr.GetConfig(),
		ServiceName:  podName,
		Namespace:    sb.Namespace,
		DatabaseName: "system",

		RunningInsideK8s:            false,
		ClientCertificateSecretName: b.Cluster().ClientTLSSecretName(),
		RootCertificateSecretName:   b.Cluster().NodeTLSSecretName(),
	}

	// Create a new database connection for the update.
	db, err := database.NewDbConnection(conn)
	require.NoError(t, err)
	defer db.Close()

	if _, err := db.Exec("CREATE DATABASE test_db"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec("USE test_db"); err != nil {
		t.Fatal(err)
	}

	// Create the "accounts" table.
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS accounts (id INT PRIMARY KEY, balance INT)"); err != nil {
		t.Fatal(err)
	}

	// Insert two rows into the "accounts" table.
	if _, err := db.Exec(
		"INSERT INTO accounts (id, balance) VALUES (1, 1000), (2, 250)"); err != nil {
		t.Fatal(err)
	}

	// Print out the balances.
	rows, err := db.Query("SELECT id, balance FROM accounts")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	t.Log("Initial balances:")
	for rows.Next() {
		var id, balance int
		if err := rows.Scan(&id, &balance); err != nil {
			t.Fatal(err)
		}
		t.Log("balances", id, balance)
	}

	countRows, err := db.Query("SELECT COUNT(*) as count FROM accounts")
	if err != nil {
		t.Fatal(err)
	}
	defer countRows.Close()
	count := getCount(t, countRows)
	if count != 2 {
		t.Fatal(fmt.Errorf("found incorrect number of rows.  Expected 2 got %v", count))
	}

}

func getCount(t *testing.T, rows *sql.Rows) (count int) {
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
	}
	return count
}
