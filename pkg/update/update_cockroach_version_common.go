/*
Copyright 2025 The Cockroach Authors

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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// makeUpdateCockroachVersionFunction takes a cockroachImage string and returns
// a function which takes a statefulset and returns the same statefulset, with
// the CockroachDB container image within changed to the new cockroachImage.
func makeUpdateCockroachVersionFunction(
	cockroachImage, version, oldVersion string,
) func(sts *v1.StatefulSet) (*v1.StatefulSet, error) {
	return func(sts *v1.StatefulSet) (*v1.StatefulSet, error) {
		timeNow := metav1.Now()
		if val, ok := sts.Annotations[resource.CrdbHistoryAnnotation]; !ok {
			sts.Annotations[resource.CrdbHistoryAnnotation] = fmt.Sprintf("%s=%s", timeNow.Format(time.RFC3339), oldVersion)
		} else {
			sts.Annotations[resource.CrdbHistoryAnnotation] = fmt.Sprintf("%s %s=%s", val, timeNow.Format(time.RFC3339), oldVersion)
		}
		sts.Annotations[resource.CrdbVersionAnnotation] = version
		sts.Annotations[resource.CrdbContainerImageAnnotation] = cockroachImage
		for i := range sts.Spec.Template.Spec.Containers {
			container := &sts.Spec.Template.Spec.Containers[i]
			// TODO "db" is hardcoded.  Make this a value in statefulset resource
			// so that we are sharing values here
			if container.Name == "db" {
				container.Image = cockroachImage
				return sts, nil
			}
		}
		return nil, errors.New("cockroachdb container not found in sts")
	}
}

// TODO some of these should probably be panics or cancel the update at least
// If we cannot find the Pod, or if we cannot find the container.
// We need to return a status and an error instead of just an error.

// makeWaitUntilCRDBPodIsRunningNewVersionFunction takes a cockroachImage and
// returns a function which takes a Kubernetes clientset, a statefulset, and a
// pod number within the statefulset. This function checks if the specified
// pod is running the new cockroachImage version and is in a `ready` state.
func makeIsCRBPodIsRunningNewVersionFunction(
	cockroachImage string,
) func(update *UpdateSts, podNumber int, l logr.Logger) error {
	return func(update *UpdateSts, podNumber int, l logr.Logger) error {
		sts := update.sts
		stsName := sts.Name
		stsNamespace := sts.Namespace
		podName := fmt.Sprintf("%s-%d", stsName, podNumber)
		clientset := update.clientset

		crdbPod, err := clientset.CoreV1().Pods(stsNamespace).Get(update.ctx, podName, metav1.GetOptions{})
		if k8sErrors.IsNotFound(err) { // this is not an error
			l.V(int(zapcore.DebugLevel)).Info("cannot find Pod", "podName", podName, "namespace", stsNamespace)
			return err
		} else if statusError, isStatus := err.(*k8sErrors.StatusError); isStatus { // this is an error
			l.Error(statusError, fmt.Sprintf("status error getting pod %v", statusError.ErrStatus.Message))
			return statusError
		} else if err != nil { // this is an error
			// TODO uncertain why this is logging is throwing an error
			// l.Error(err, "error getting pod")
			// using this line instead of the one above since the above line is throwing an error
			// during e2e testing
			l.V(int(zapcore.ErrorLevel)).Info("error findinging Pod", "podName", podName, "namespace", stsNamespace)
			return err
		}

		for i := range crdbPod.Spec.Containers {
			container := &crdbPod.Spec.Containers[i]
			// TODO this is hard coded and resource statefulset needs to use this
			if container.Name == "db" {

				// TODO this is not an error but should return a wait status
				if container.Image != cockroachImage {
					l.V(int(zapcore.DebugLevel)).Info("Pod is not updated to current image.")
					return fmt.Errorf("%s pod is on image %s, expected %s", podName, container.Image, cockroachImage)
				}

				// TODO this is not an error but should return a wait status
				// CRDB pod is updated to new Cockroach image. Now check
				// that the pod is in a ready state before proceeding.
				if !kube.IsPodReady(crdbPod) {
					l.V(int(zapcore.DebugLevel)).Info("Pod is not ready yet.", "podName", podName, "stsName", stsName)
					return fmt.Errorf("%s pod not ready yet", crdbPod)
				}

				l.V(int(zapcore.DebugLevel)).Info("is running new version on", "podName", podName, "stsName", stsNamespace)
				return nil
			}
		}

		// TODO how do we even get here??
		// I am not certain that this code is even used.

		err = fmt.Errorf("cockroachdb container not found within the cockroach pod")
		l.Error(err, "container not found")
		return err
	}
}

// Note that while CockroachDB considers 19.2 to be a major release, if we follow
// semantic versioning (https://semver.org/spec/v2.0.0.html), both 19.1 and 19.2
// is a minor release of version 19. The code below parses the version as if it
// were following the semantic versioning spec.
func isPatch(wantVersion *semver.Version, currentVersion *semver.Version) bool {
	return currentVersion.Major() == wantVersion.Major() && currentVersion.Minor() == wantVersion.Minor()
}

// getReleaseType returns if a release is Innovative or Regular.
func getReleaseType(major, minor int) ReleaseType {
	// Before 25.1, we define them explicitly
	switch fmt.Sprintf("%d.%d", major, minor) {
	case "24.1", "24.3":
		return Regular
	case "24.2":
		return Innovative
	}

	// Post 25.1: Odd releases (1,3) are Innovative, Even (2,4) are Regular
	if minor%2 == 1 {
		return Innovative
	}
	return Regular
}

// generateReleases generates a list of release versions from a 24.1 to the specified year.
// The releases follow a "YY.Q" format, where YY represents the last two digits of the year, and Q represents the quarter.
//
// Parameters:
// - upToYear: The last two digits of the year up to which releases should be generated.
//
// Returns:
//   - A slice of strings containing release versions from "24.1" onward, up to the specified year, with four
//     releases per year (except 2024).
func generateReleases(upToYear int) []string {
	var releases = []string{"24.1", "24.2", "24.3"}

	for year := 25; year <= upToYear; year++ {
		for quarter := 1; quarter <= 4; quarter++ {
			releases = append(releases, fmt.Sprintf("%d.%d", year, quarter))
		}
	}

	return releases
}

// getNextReleases returns a list of possible upgradable release versions following the given currentVersion.
// It iterates through the generated release versions for the current year and collects the next releases
// after finding the currentVersion. The function stops collecting releases once it encounters the next regular release.
//
// Parameters:
// - currentVersion: The current release version in the format "YY.Q" (e.g., "24.1").
//
// Returns:
// - A slice of strings containing the possible upgradable release versions, stopping at the next regular release.
func getNextReleases(currentVersion string) []string {
	var nextReleases []string
	var found bool

	releases := generateReleases(time.Now().Year() % 100)
	for _, release := range releases {
		year, _ := strconv.Atoi(strings.Split(release, ".")[0])
		quarter, _ := strconv.Atoi(strings.Split(release, ".")[1])

		if found {
			nextReleases = append(nextReleases, release)
			if getReleaseType(year, quarter) == Regular {
				break // Stop at the next regular release
			}
		}
		if release == currentVersion {
			found = true
		}
	}

	return nextReleases
}

// getPreviousReleases returns a list of possible rollback release versions preceding the given currentVersion.
// It iterates through the generated release versions in reverse order and collects previous releases
// until it encounters the last regular release before the currentVersion.
//
// Parameters:
// - currentVersion: The current release version in the format "YY.Q" (e.g., "24.1").
//
// Returns:
// - A slice of strings containing possible rollback release versions, stopping at the last regular release before the currentVersion.
func getPreviousReleases(currentVersion string) []string {
	var prevReleases []string
	var found bool

	releases := generateReleases(time.Now().Year() % 100)
	for i := len(releases) - 1; i >= 0; i-- {
		release := releases[i]
		year, _ := strconv.Atoi(strings.Split(release, ".")[0])
		quarter, _ := strconv.Atoi(strings.Split(release, ".")[1])
		if found {
			prevReleases = append(prevReleases, release)
			if getReleaseType(year, quarter) == Regular {
				break
			}
		}
		if release == currentVersion {
			found = true
		}
	}

	return prevReleases
}

// isMajorUpgradeAllowed determines whether an upgrade from the current version to the desired version is allowed
// based on the release cycle and upgrade policies.
//
// Upgrade Rules:
// 1. For versions prior to 2024:
//   - A major upgrade is allowed if
//   - The major version remains the same and the minor version increments by 1 (e.g., 19.1 -> 19.2).
//   - The major version increments by 1 and the minor version resets to 1. (e.g., 19.2 -> 20.1).
//
// 2. For version of 2024:
//   - There are 3 releases (24.1, 24.2, 24.3) where 24.2 being innovative release
//
// 3. For versions from 2025 onwards:
//   - Releases follow a quarterly cycle, with some releases designated as innovative.
//   - Users are allowed to skip upgrading to an innovative release.
//   - The upgrade is permitted if the desired version matches one of the next valid releases.
//
// Parameters:
// - wantVersion: The target version to which an upgrade is requested.
// - currentVersion: The currently installed version.
//
// Returns:
// - true if the upgrade is allowed based on the release rules; false otherwise.
func isMajorUpgradeAllowed(wantVersion *semver.Version, currentVersion *semver.Version) bool {
	// Two cases:
	// 19.1 to 19.2 -> same year
	// 19.2 to 20.1 -> next year

	// Since 2025, we have adopted a quarterly release cycle, with two of the four annual releases designated
	// as innovative releases. Users have the option to skip upgrading to an innovative release.
	// For 2024, we just had 3 releases overall, with 24.2 as the innovation release.
	if currentVersion.Major() >= 24 {
		// Four Cases:
		// 24.1 to 24.2 -> Same year without skipping innovative release
		// 24.1 to 24.3 -> Same year with skipping innovative release
		// 24.3 to 25.1 -> Next year without skipping innovative release
		// 24.3 to 25.2 -> Next year with skipping innovative release
		nextPossibleRelease := getNextReleases(fmt.Sprintf("%d.%d", currentVersion.Major(), currentVersion.Minor()))
		for _, version := range nextPossibleRelease {
			if version == fmt.Sprintf("%d.%d", wantVersion.Major(), wantVersion.Minor()) {
				return true
			}
		}

		return false
	}

	return (currentVersion.Major() == wantVersion.Major() && currentVersion.Minor()+1 == wantVersion.Minor()) ||
		(currentVersion.Major()+1 == wantVersion.Major() && currentVersion.Minor()-1 == wantVersion.Minor())
}

// isMajorRollbackAllowed determines whether rolling back from the current version to the desired version is allowed
// based on the release cycle and rollback policies.
//
// Rollback Rules:
// 1. For versions prior to 2024:
//   - A rollback is allowed if:
//   - The major version remains the same and the minor version decrements by 1 (e.g., 19.2 -> 19.1).
//   - The major version decrements by 1 and the minor version is the last release. (e.g., 20.1 -> 19.2).
//
// 2. For versions from 2024 onwards:
//   - Users can skip rolling back to an innovative release.
//   - The rollback is permitted if the desired version matches one of the valid previous releases.
//
// Parameters:
// - wantVersion: The target version to which a rollback is requested.
// - currentVersion: The currently installed version.
//
// Returns:
// - true if the rollback is allowed based on the rollback policies; false otherwise.
func isMajorRollbackAllowed(wantVersion *semver.Version, currentVersion *semver.Version) bool {
	// Two cases:
	// 19.2 to 19.1 -> same year
	// 20.1 to 19.2 -> previous year

	// Since 2024, users have the option to skip rollback to an innovative release.
	if wantVersion.Major() >= 24 {
		// Four cases:
		// 24.2 -> 24.1 -> Same year without skipping innovative release
		// 24.3 -> 24.1 -> Same year with skipping innovative release
		// 25.1 -> 24.3 -> Previous year without skipping innovative release
		// 25.2 -> 24.3 -> Previous year with skipping innovative release
		rollbackReleases := getPreviousReleases(fmt.Sprintf("%d.%d", currentVersion.Major(), currentVersion.Minor()))
		for _, version := range rollbackReleases {
			if version == fmt.Sprintf("%d.%d", wantVersion.Major(), wantVersion.Minor()) {
				return true
			}
		}

		return false
	}

	return (currentVersion.Major() == wantVersion.Major() && currentVersion.Minor() == wantVersion.Minor()+1) ||
		(currentVersion.Major() == wantVersion.Major()+1 && currentVersion.Minor() == wantVersion.Minor()-1)
}
