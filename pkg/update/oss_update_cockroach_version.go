package update

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"database/sql"

	"github.com/Masterminds/semver"
	"github.com/cenkalti/backoff"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// This file is refactoring update_cockroach_version.go into
// funcs that do not rely on external crl code.
// Once this file is opensource we can remove all of the oss_ prefixes on the
// funcs and variables.

var oss_validPreserveDowngradeOptionSetting = regexp.MustCompile(`^[0-9][0-9]\.[0-9]$`) // e.g. 19.2

type OSS_UpdateNotAllowed struct {
	cur      *semver.Version
	want     *semver.Version
	preserve *semver.Version
	extra    string
}

var _ error = OSS_UpdateNotAllowed{}

func (e OSS_UpdateNotAllowed) Error() string {
	return fmt.Sprintf("upgrading from %v to %v with preserve downgrade option set to %v not allowed: %s", e.cur, e.want, e.preserve, e.extra)
}

// TODO update docs

// updateClusterCockroachVersion is the private version of
// UpdateClusterCockroachVersion, and allows specifying custom pod timeouts,
// among other things, in order to enable unit testing.
func oss_updateClusterCockroachVersion(
	ctx context.Context,
	cockroachVersion string,
	clientset kubernetes.Interface,
	wantImageName string,
	wantVersion *semver.Version,
	podUpdateTimeout time.Duration,
	podMaxPollingInterval time.Duration,
	db *sql.DB,
	sleeper Sleeper,
	stsName string,
	ns string,
) error {
	k, err := oss_kindAndCheckPreserveDowngradeSetting(ctx, wantVersion, cockroachVersion, db)
	if err != nil {
		return err
	}

	// TODO(josh): Either:
	//  1. Delete CockroachVersion from DB. Always fetch from k8s.
	//  2. Check that CockroachVersion in DB is equal to version in k8s.
	currentVersion, err := semver.NewVersion(cockroachVersion)
	if err != nil {
		return errors.Wrapf(err, "parsing current version failed")
	}

	l := fromContext(ctx).With(
		zap.String("kind", k),
		zap.String("from", cockroachVersion),
		zap.String("to", wantImageName+":"+wantVersion.Original()),
	)
	l.Info("starting upgrade")

	// TODO we need to move this into oss_kindAndCheckPreserveDowngradeSetting
	if isForwardOneMajorVersion(wantVersion, currentVersion) {
		if err := oss_setDowngradeOption(ctx, wantVersion, currentVersion, db, l); err != nil {
			return errors.Wrapf(err, "setting downgrade option for major roll forward failed")
		}
	}

	wantImage := fmt.Sprintf("%s:%s", wantImageName, wantVersion.Original())

	updateFunction := makeUpdateCockroachVersionFunction(wantImage)
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

	return oss_updateClusterStatefulSets(ctx, clientset, updateSuite, podUpdateTimeout, podMaxPollingInterval, sleeper, l, stsName, ns)
}

// updateClusterStatefulSets takes a context, a cluster, a vault client, and an
// updateFunctionSuite. It uses the functions within the updateFunctionSuite to
// update the CockroachDB StatefulSet in each region of a CockroachDB cluster.
func oss_updateClusterStatefulSets(
	ctx context.Context,
	clientset kubernetes.Interface,
	updateSuite *updateFunctionSuite,
	podUpdateTimeout time.Duration,
	podMaxPollingInterval time.Duration,
	sleeper Sleeper,
	l *zap.Logger,
	stsName string,
	ns string,
) error {
	// TODO see what skipSleep should be doing here
	// It is the first param returned by UpdateClusterRegionStatefulSet
	_, err := UpdateClusterRegionStatefulSet(
		ctx,
		clientset,
		ns,
		stsName,
		updateSuite,
		oss_makeWaitUntilAllPodsReadyFunc(clientset, podUpdateTimeout, podMaxPollingInterval, stsName, ns),
		podUpdateTimeout,
		podMaxPollingInterval,
		sleeper,
		l)
	if err != nil {
		return err
	}
	return nil
}

// waitUntilAllPodsReady waits until all pods in all statefulsets are in the
// ready state. The ready state implies all nodes are passing node liveness.
func oss_makeWaitUntilAllPodsReadyFunc(
	clientset kubernetes.Interface,
	podUpdateTimeout time.Duration,
	maxPodPollingInterval time.Duration,
	stsName string,
	ns string,
) func(ctx context.Context, l *zap.Logger) error {
	return func(ctx context.Context, l *zap.Logger) error {

		l.Info("waiting until all pods are in the ready state")
		f := func() error {

			sts, err := clientset.AppsV1().StatefulSets(ns).Get(ctx, stsName, metav1.GetOptions{})
			if err != nil {
				l.Warn("could not find crdb sts")
				return errors.Wrapf(err, "could not find crdb sts for %s", stsName)
			}
			got := int(sts.Status.ReadyReplicas)
			// TODO need to test this
			// we could also use the number of pods defined by the operator
			numCRDBPods := int(sts.Status.Replicas)
			if got != numCRDBPods {
				l.Sugar().Warnf("number of ready replicas is %v, not equal to num CRDB pods %v", got, numCRDBPods)
				return fmt.Errorf("number of ready replicas is %v, not equal to num CRDB pods %v", got, numCRDBPods)
			}

			l.Info("all replicas are ready")
			return nil
		}

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = podUpdateTimeout
		b.MaxInterval = maxPodPollingInterval
		return backoff.Retry(f, b)
	}
}

func oss_kindAndCheckPreserveDowngradeSetting(ctx context.Context, wantVersion *semver.Version, cockroachVersion string, db *sql.DB) (string, error) {
	// TODO(josh): Either:
	//  1. Delete CockroachVersion from DB. Always fetch from k8s.
	//  2. Check that CockroachVersion in DB is equal to version in k8s.
	currentVersion, err := semver.NewVersion(cockroachVersion)
	if err != nil {
		return "UNKNOWN", errors.Wrapf(err, "parsing current version failed")
	}

	if isPatch(wantVersion, currentVersion) {
		return "PATCH", nil
	} else if isForwardOneMajorVersion(wantVersion, currentVersion) {
		s := "MAJOR_UPGRADE"
		preserve, err := oss_preserveDowngradeSetting(ctx, db)
		if err != nil {
			return s, err
		}
		// To do a roll forward, preserve downgrade option should either be
		// unset, or set to current version. If unset, kubeupdate will set it to
		// current version.
		if preserve.Compare(&semver.Version{}) != 0 &&
			(preserve.Major() != currentVersion.Major() || preserve.Minor() != currentVersion.Minor()) {
			return s, OSS_UpdateNotAllowed{
				cur:      currentVersion,
				want:     wantVersion,
				preserve: preserve,
				extra:    "can't roll forward due to preserve downgrade option",
			}
		}
		return s, nil
	} else if isBackOneMajorVersion(wantVersion, currentVersion) {
		s := "MAJOR_ROLLBACK"
		preserve, err := oss_preserveDowngradeSetting(ctx, db)
		if err != nil {
			return s, err
		}
		// To do a rollback, preserve downgrade option must be set to the major
		// version to which kubeupdate is rolling back.
		if preserve.Major() != wantVersion.Major() || preserve.Minor() != wantVersion.Minor() {
			return s, OSS_UpdateNotAllowed{
				cur:      currentVersion,
				want:     wantVersion,
				preserve: preserve,
				extra:    "can't rollback since release already finalized",
			}
		}
		return s, nil
	}
	return "UNKNOWN", OSS_UpdateNotAllowed{
		cur:   currentVersion,
		want:  wantVersion,
		extra: "only patches, rolling forward one major version, & rolling back one major version supported",
	}
}

func oss_preserveDowngradeSetting(ctx context.Context, db *sql.DB) (*semver.Version, error) {
	preserveDowngradeSetting, err := getClusterSetting(ctx, db, "cluster.preserve_downgrade_option")
	if err != nil {
		return nil, errors.Wrapf(err, "getting preserve downgrade option failed")
	}
	if preserveDowngradeSetting == "" {
		return &semver.Version{}, nil // empty semver.Version means unset preserve downgrade option
	}
	if !oss_validPreserveDowngradeOptionSetting.MatchString(preserveDowngradeSetting) {
		return nil, fmt.Errorf("%s is not a valid preserve downgrade option setting", preserveDowngradeSetting)
	}
	preserveDowngradeVersion, err := semver.NewVersion("v" + preserveDowngradeSetting + ".0") // e.g. 19.2
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse finalization flag")
	}
	return preserveDowngradeVersion, nil
}

func oss_setDowngradeOption(ctx context.Context, wantVersion *semver.Version, currentVersion *semver.Version, db *sql.DB, l *zap.Logger) error {
	newDowngradeOption := fmt.Sprintf("%d.%d", currentVersion.Major(), currentVersion.Minor())
	if !oss_validPreserveDowngradeOptionSetting.MatchString(newDowngradeOption) {
		return fmt.Errorf("%s is not a valid preserve downgrade option setting", newDowngradeOption)
	}
	if err := setClusterSetting(ctx, db, PreserveDowngradeOptionClusterSetting, newDowngradeOption); err != nil {
		return errors.Wrapf(err, "setting preserve downgrade option failed")
	}

	l.Info("setting downgrade option since major version", zap.String("cluster.preserve_downgrade_option", newDowngradeOption))

	return nil
}

// TODO move to its own file in OSS
type loggerKey struct{}

// FromContext returns either a logger instance attached to `ctx` or
// the global logger with the field "missingFromCtx" set to true.
//
// If possible, logger should be threaded down rather than being passed
// via context. However, it is more important to have logging than
// perfectly clean code
func fromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return logger
	}

	return zap.L().With(zap.Bool("missingFromCtx", true))
}

// roleMembership represents role membership for a particular database user.
type roleMembership struct {
	// Name is the name of the role membership.
	Name string

	// IsAdmin represents whether the "WITH ADMIN OPTION" is granted to the
	// user's role. Enabling this option allows the user to grant or revoke
	// membership of the associated role to other users.
	IsAdmin bool
}

// ListRoleGrantsForUser returns a list of role memberships for the given user.
func listRoleGrantsForUser(
	ctx context.Context, db *sql.DB, username string,
) ([]roleMembership, error) {

	query := fmt.Sprintf(`SHOW GRANTS ON ROLE FOR %s`, pq.QuoteIdentifier(username))
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []roleMembership
	for rows.Next() {
		var role roleMembership
		// `member` is ignored here because we're querying for the same user.
		// Ideally, this function should take in a list of users, and return
		// `map[string][]roleMembership`.
		var member string

		if err := rows.Scan(&role.Name, &member, &role.IsAdmin); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func grantRoleToUser(ctx context.Context, db *sql.DB, role roleMembership, username string) error {

	query := fmt.Sprintf("GRANT %s TO %s", role.Name, pq.QuoteIdentifier(username))
	if role.IsAdmin {
		query += " WITH ADMIN OPTION"
	}
	if _, err := db.ExecContext(ctx, query); err != nil {
		return err
	}
	return nil
}

func getClusterSetting(ctx context.Context, db *sql.DB, name string) (string, error) {
	// TODO do we need this?
	/*
		if err := isValidClusterSettingName(name); err != nil {
			return "", err
		}
	*/

	r := db.QueryRowContext(ctx, fmt.Sprintf("SHOW CLUSTER SETTING %s", name))
	var value string
	if err := r.Scan(&value); err != nil {
		return "", errors.Wrapf(err, "failed to get %s", name)
	}
	return value, nil
}

func setClusterSetting(ctx context.Context, db *sql.DB, name string, value string) error {
	/*
		if err := isValidClusterSettingName(name); err != nil {
			return err
		}
	*/

	sqlStr := fmt.Sprintf("SET CLUSTER SETTING %s = $1", name)
	if _, err := db.Exec(sqlStr, value); err != nil {
		return errors.Wrapf(err, "failed to set %s to %s", name, value)
	}
	return nil
}
