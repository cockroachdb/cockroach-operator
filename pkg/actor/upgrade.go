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
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newUpgrade(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &upgrade{
		action: newAction("upgrade", scheme, cl),
		config: config,
	}
}

// upgrade handles minor and major version upgrades without finalization
type upgrade struct {
	action

	config *rest.Config
}

func (up *upgrade) Handles(conds []api.ClusterCondition) bool {
	return condition.False(api.NotInitializedCondition, conds) && utilfeature.DefaultMutableFeatureGate.Enabled(features.Upgrade)
}

func (up *upgrade) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := up.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.Info("checking upgrade opportunities")

	stsName := cluster.StatefulSetName()

	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      stsName,
	}
	statefulSet := &appsv1.StatefulSet{}
	if err := up.client.Get(ctx, key, statefulSet); err != nil {
		return errors.Wrap(err, "failed to fetch statefulset")
	}

	if statefulSetIsUpdating(statefulSet) {
		return NotReadyErr{Err: errors.New("statefulset is updating, waiting for the update to finish")}
	}

	dbContainer, err := kube.FindContainer(resource.DbContainerName, &statefulSet.Spec.Template.Spec)
	if err != nil {
		return err
	}

	// nothing to be done
	if dbContainer.Image == cluster.Spec().Image.Name {
		log.Info("no version changes needed")
		return nil
	}

	specStr := getVersionFromImage(cluster.Spec().Image.Name)
	if specStr == "" {
		return fmt.Errorf("Unknown CockroachDB version in spec: %s", cluster.Spec().Image.Name)
	}

	specVer, err := semver.NewVersion(specStr)
	if err != nil {
		return errors.Wrapf(err, "failed to parse spec image version: %s", specStr)
	}

	clusterStr, err := up.getClusterSetting(cluster, "version")
	if err != nil {
		return errors.Wrapf(err, "failed to fetch cluster version")
	}

	// Caret range comparison `^` to detect major change
	// https://github.com/Masterminds/semver#caret-range-comparisons-major)
	contraintStr := fmt.Sprintf("^%s", clusterStr)
	constraint, err := semver.NewConstraint(contraintStr)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize version contaraint: %s", contraintStr)
	}

	log.Info("detected versions", "spec", specVer.String(), "cluster", clusterStr)

	// Major upgrade
	if !constraint.Check(specVer) {
		log.Info("preserving downgrade option")
		if err := up.preserveDowngrade(cluster, clusterStr); err != nil {
			return err
		}
	}

	update := statefulSet.DeepCopy()
	updatedContainer, _ := kube.FindContainer(resource.DbContainerName, &update.Spec.Template.Spec)
	updatedContainer.Image = cluster.Spec().Image.Name

	if err := up.client.Update(ctx, update); err != nil {
		return errors.Wrap(err, "failed to update image version")
	}

	log.Info("scheduled update", "new version", specStr)
	CancelLoop(ctx)
	return nil
}

// TODO this does not handle sha's
// TODO we are running into the same problem where the version of
// CR is parsed from the image tag name, we need to fix this

func getImageNameNoVersion(image string) string {
	i := strings.LastIndex(image, ":")
	if i == -1 {
		return image
	}

	return image[:i]
}

func getVersionFromImage(image string) string {
	i := strings.LastIndex(image, ":")
	if i == -1 {
		return ""
	}

	return image[i+1:]
}

func (up *upgrade) getClusterSetting(cluster *resource.Cluster, setting string) (string, error) {
	rows, err := up.showClusterSetting(cluster, setting)
	if err != nil {
		return "", err
	}

	if len(rows) == 1 {
		// we only have the header.
		return "", nil
	}

	if len(rows) != 2 || len(rows[0]) != 1 || len(rows[1]) != 1 {
		return "", fmt.Errorf("expected two rows with 1 column, got: %v", rows)
	}

	if rows[0][0] != setting {
		return "", fmt.Errorf("expected column named \"%s\", got: %s", setting, rows[0][0])
	}

	return rows[1][0], nil
}

func (up *upgrade) showClusterSetting(cluster *resource.Cluster, setting string) ([][]string, error) {
	cmd := []string{
		"/bin/bash",
		"-c",
		"/cockroach/cockroach sql " + cluster.SecureMode() +
			fmt.Sprintf(" -e 'SHOW CLUSTER SETTING %s' --format csv", setting),
	}

	stdout, stderr, err := kube.ExecInPod(up.scheme, up.config, cluster.Namespace(),
		fmt.Sprintf("%s-0", cluster.StatefulSetName()), resource.DbContainerName, cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch cluster setting %s: %s", setting, stderr)
	}

	return csv.NewReader(strings.NewReader(stdout)).ReadAll()
}

func statefulSetIsUpdating(ss *appsv1.StatefulSet) bool {
	if ss.Status.ObservedGeneration == 0 {
		return false
	}

	if ss.Status.CurrentRevision != ss.Status.UpdateRevision {
		return true
	}

	if ss.Generation > ss.Status.ObservedGeneration && *ss.Spec.Replicas == ss.Status.Replicas {
		return true
	}

	return false
}

func (up *upgrade) preserveDowngrade(cluster *resource.Cluster, clusterVer string) error {
	cmd := []string{
		"/bin/bash",
		"-c",
		"/cockroach/cockroach sql " + cluster.SecureMode() +
			fmt.Sprintf(" -e \"SET CLUSTER SETTING cluster.preserve_downgrade_option = '%s'\"", clusterVer),
	}

	_, stderr, err := kube.ExecInPod(up.scheme, up.config, cluster.Namespace(),
		fmt.Sprintf("%s-0", cluster.StatefulSetName()), resource.DbContainerName, cmd)
	if err != nil {
		return errors.Wrapf(err, "failed to update cluster.preserve_downgrade_option: %s", stderr)
	}

	return nil
}
