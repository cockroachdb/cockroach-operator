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

package update

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"database/sql"

	"github.com/Masterminds/semver/v3"
	"github.com/cenkalti/backoff"
	"github.com/cockroachdb/cockroach-operator/pkg/clustersql"
	"github.com/cockroachdb/cockroach-operator/pkg/healthchecker"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var validPreserveDowngradeOptionSetting = regexp.MustCompile(`^[0-9][0-9]\.[0-9]$`) // e.g. 19.2

type UpdateNotAllowed struct {
	cur      *semver.Version
	want     *semver.Version
	preserve *semver.Version
	extra    string
}

var _ error = UpdateNotAllowed{}

func (e UpdateNotAllowed) Error() string {
	return fmt.Sprintf("upgrading from %v to %v with preserve downgrade option set to %v not allowed: %s", e.cur, e.want, e.preserve, e.extra)
}

type UpdateRoach struct {
	CurrentVersion *semver.Version
	WantVersion    *semver.Version
	WantImageName  string
	StsName        string
	StsNamespace   string
	Db             *sql.DB
}

type UpdateCluster struct {
	Clientset             kubernetes.Interface
	PodUpdateTimeout      time.Duration
	PodMaxPollingInterval time.Duration
	HealthChecker         healthchecker.HealthChecker
}

// UpdateClusterCockroachVersion, and allows specifying custom pod timeouts,
// among other things, in order to enable unit testing.
func UpdateClusterCockroachVersion(
	ctx context.Context,
	update *UpdateRoach,
	cluster *UpdateCluster,
	l logr.Logger,
) error {

	l.WithValues(
		"from",
		update.CurrentVersion.Original(),
		"to",
		update.WantImageName+":"+update.WantVersion.Original(),
	)

	kind, err := kindAndCheckPreserveDowngradeSetting(ctx, update.WantVersion, update.CurrentVersion, update.Db, l)
	if err != nil {
		return err
	}
	l.WithValues(
		"kind",
		kind,
	)

	l.V(int(zapcore.InfoLevel)).Info("starting upgrade")

	if isForwardOneMajorVersion(update.WantVersion, update.CurrentVersion) {
		if err := setDowngradeOption(ctx, update.WantVersion, update.CurrentVersion, update.Db, l); err != nil {
			return errors.Wrapf(err, "setting downgrade option for major roll forward failed")
		}
	}
	var wantImage string = update.WantImageName
	//if the image is already in the sha256 format  we keep it as it is
	//otherwise we build the concatenate the image with the version to
	//obtain the correct format
	if !strings.Contains(update.WantImageName, "@sha256") {
		wantImage = fmt.Sprintf("%s:%s", update.WantImageName, update.WantVersion.Original())
	}

	updateFunction := makeUpdateCockroachVersionFunction(wantImage, update.WantVersion.Original(), update.CurrentVersion.Original())
	perPodVerificationFunction := makeIsCRBPodIsRunningNewVersionFunction(
		wantImage,
	)
	updateStrategyFunction := PartitionedRollingUpdateStrategy(
		perPodVerificationFunction,
	)

	updateSuite := &updateFunctionSuite{
		updateFunc:         updateFunction,
		updateStrategyFunc: updateStrategyFunction,
	}

	return updateClusterStatefulSets(ctx, update, cluster, updateSuite, l)
}

// updateClusterStatefulSets takes a context, a cluster, a vault client, and an
// updateFunctionSuite. It uses the functions within the updateFunctionSuite to
// update the CockroachDB StatefulSet in each region of a CockroachDB cluster.
func updateClusterStatefulSets(
	ctx context.Context,
	update *UpdateRoach,
	cluster *UpdateCluster,
	updateSuite *updateFunctionSuite,
	l logr.Logger,
) error {
	// TODO see what skipSleep should be doing here
	// It is the first param returned by UpdateClusterRegionStatefulSet
	_, err := UpdateClusterRegionStatefulSet(
		ctx,
		cluster.Clientset,
		update.StsName,
		update.StsNamespace,
		updateSuite,
		makeWaitUntilAllPodsReadyFunc(ctx, cluster, update),
		cluster.PodUpdateTimeout,
		cluster.PodMaxPollingInterval,
		cluster.HealthChecker,
		l)
	if err != nil {
		return err
	}
	return nil
}

// waitUntilAllPodsReady waits until all pods in all statefulsets are in the
// ready state. The ready state implies all nodes are passing node liveness.
func makeWaitUntilAllPodsReadyFunc(
	ctx context.Context,
	cluster *UpdateCluster,
	update *UpdateRoach,
) func(ctx context.Context, l logr.Logger) error {
	return func(ctx context.Context, l logr.Logger) error {

		l.V(int(zapcore.DebugLevel)).Info("waiting until all pods are in the ready state")
		f := func() error {

			sts, err := cluster.Clientset.AppsV1().StatefulSets(update.StsNamespace).Get(ctx, update.StsName, metav1.GetOptions{})
			if err != nil {
				return handleStsError(err, l, update.StsName, update.StsNamespace)
			}
			got := int(sts.Status.ReadyReplicas)
			// TODO need to test this
			// we could also use the number of pods defined by the operator
			numCRDBPods := int(sts.Status.Replicas)
			if got != numCRDBPods {
				l.Error(err, fmt.Sprintf("number of ready replicas is %v, not equal to num CRDB pods %v", got, numCRDBPods))
				return err
			}

			l.V(int(zapcore.DebugLevel)).Info("all replicas are ready makeWaitUntilAllPodsReadyFunc update_cockroach_version.go")
			return nil
		}

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = cluster.PodUpdateTimeout
		b.MaxInterval = cluster.PodMaxPollingInterval
		return backoff.Retry(f, b)
	}
}

func kindAndCheckPreserveDowngradeSetting(
	ctx context.Context,
	wantVersion *semver.Version,
	currentVersion *semver.Version,
	db *sql.DB,
	l logr.Logger,
) (string, error) {
	// TODO(josh): Either:
	//  1. Delete CockroachVersion from DB. Always fetch from k8s.
	//  2. Check that CockroachVersion in DB is equal to version in k8s.
	//currentVersion, err := semver.NewVersion(cockroachVersion)
	//if err != nil {
	//	return "UNKNOWN", errors.Wrapf(err, "parsing current version failed")
	//}

	if isPatch(wantVersion, currentVersion) {
		l.V(int(zapcore.DebugLevel)).Info("patch upgrade")
		return "PATCH", nil
	} else if isForwardOneMajorVersion(wantVersion, currentVersion) {
		l.V(int(zapcore.DebugLevel)).Info("major upgrade")
		s := "MAJOR_UPGRADE"
		preserve, err := preserveDowngradeSetting(ctx, db)
		if err != nil {
			return s, err
		}
		// To do a roll forward, preserve downgrade option should either be
		// unset, or set to current version. If unset, kubeupdate will set it to
		// current version.
		if (preserve.Compare(&semver.Version{}) != 0 &&
			(preserve.Major() != currentVersion.Major() || preserve.Minor() != currentVersion.Minor())) {
			return s, UpdateNotAllowed{
				cur:      currentVersion,
				want:     wantVersion,
				preserve: preserve,
				extra:    "can't roll forward due to preserve downgrade option",
			}
		}
		return s, nil
	} else if isBackOneMajorVersion(wantVersion, currentVersion) {
		l.V(int(zapcore.DebugLevel)).Info("major rollback")
		s := "MAJOR_ROLLBACK"
		preserve, err := preserveDowngradeSetting(ctx, db)
		if err != nil {
			return s, err
		}
		// To do a rollback, preserve downgrade option must be set to the major
		// version to which kubeupdate is rolling back.
		if preserve.Major() != wantVersion.Major() || preserve.Minor() != wantVersion.Minor() {
			return s, UpdateNotAllowed{
				cur:      currentVersion,
				want:     wantVersion,
				preserve: preserve,
				extra:    "can't rollback since release already finalized",
			}
		}
		return s, nil
	}

	err := UpdateNotAllowed{
		cur:   currentVersion,
		want:  wantVersion,
		extra: "only patches, rolling forward one major version, & rolling back one major version supported",
	}
	l.Error(err, "unknown upgrade")
	return "UNKNOWN", err
}

func preserveDowngradeSetting(ctx context.Context, db *sql.DB) (*semver.Version, error) {
	preserveDowngradeSetting, err := clustersql.GetClusterSetting(ctx, db, "cluster.preserve_downgrade_option")
	if err != nil {
		return nil, errors.Wrapf(err, "getting preserve downgrade option failed")
	}
	if preserveDowngradeSetting == "" {
		return &semver.Version{}, nil // empty semver.Version means unset preserve downgrade option
	}
	if !validPreserveDowngradeOptionSetting.MatchString(preserveDowngradeSetting) {
		return nil, fmt.Errorf("%s is not a valid preserve downgrade option setting", preserveDowngradeSetting)
	}
	preserveDowngradeVersion, err := semver.NewVersion("v" + preserveDowngradeSetting + ".0") // e.g. 19.2
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse finalization flag")
	}
	return preserveDowngradeVersion, nil
}

func setDowngradeOption(ctx context.Context, wantVersion *semver.Version, currentVersion *semver.Version, db *sql.DB, l logr.Logger) error {
	newDowngradeOption := fmt.Sprintf("%d.%d", currentVersion.Major(), currentVersion.Minor())
	if !validPreserveDowngradeOptionSetting.MatchString(newDowngradeOption) {
		return fmt.Errorf("%s is not a valid preserve downgrade option setting", newDowngradeOption)
	}
	if err := clustersql.SetClusterSetting(ctx, db, PreserveDowngradeOptionClusterSetting, newDowngradeOption); err != nil {
		return errors.Wrapf(err, "setting preserve downgrade option failed")
	}

	l.V(int(zapcore.DebugLevel)).Info("set downgrade option since major version", "cluster.preserve_downgrade_option", newDowngradeOption)

	return nil
}
