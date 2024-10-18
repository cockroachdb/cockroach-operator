/*
Copyright 2024 The Cockroach Authors

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

package clustersql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dustin/go-humanize"
)

// ErrInvalidClusterSettingName is returned when the supplied setting name isn't valid.
var ErrInvalidClusterSettingName = fmt.Errorf("only letters, numbers, underscores, and periods are allowed")

// Cluster names have letters, underscores, and periods. See here for more:
// https://www.cockroachlabs.com/docs/stable/cluster-settings.html
var validClusterSettingNameRE = regexp.MustCompile(`^[a-z_.\d]+$`)

func validateSettingName(name string) error {
	if !validClusterSettingNameRE.MatchString(name) {
		return errors.Wrapf(ErrInvalidClusterSettingName, "%s is not a valid cluster setting", name)
	}
	return nil
}

// GetClusterSetting func
func GetClusterSetting(ctx context.Context, db *sql.DB, name string) (string, error) {
	if err := validateSettingName(name); err != nil {
		return "", err
	}

	r := db.QueryRowContext(ctx, fmt.Sprintf("SHOW CLUSTER SETTING %s", name))
	var value string
	if err := r.Scan(&value); err != nil {
		return "", errors.Wrapf(err, "failed to get %s", name)
	}
	return value, nil
}

// SetClusterSetting func
func SetClusterSetting(ctx context.Context, db *sql.DB, name, value string) error {
	if err := validateSettingName(name); err != nil {
		return err
	}

	sql := fmt.Sprintf("SET CLUSTER SETTING %s = $1", name)
	if _, err := db.Exec(sql, value); err != nil {
		return errors.Wrapf(err, "failed to set %s to %s", name, value)
	}
	return nil
}

// RangeMoveDuration calculates the slowest time.Duration that a range would
// reasonably take to move from one node to another.
// This duration does not account for IOPs or cluster load. If used as a timeout
// a multiple of this value should be used.
func RangeMoveDuration(ctx context.Context, db *sql.DB, zones ...Zone) (time.Duration, error) {
	rebalanceRate, err := GetClusterSetting(ctx, db, "kv.snapshot_rebalance.max_rate")
	if err != nil {
		return 0, errors.Wrap(err, "failed to get kv.snapshot_rebalance.max_rate")
	}

	recoveryRate, err := GetClusterSetting(ctx, db, "kv.snapshot_recovery.max_rate")
	if err != nil {
		// This setting has been removed in 24.1, so if an error is returned set it to a default
		// huge number
		recoveryRate = "10TB"
	}

	rebalanceBytes, err := humanize.ParseBytes(rebalanceRate)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse kv.snapshot_rebalance.max_rate as uint64")
	}

	recoveryBytes, err := humanize.ParseBytes(recoveryRate)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse kv.snapshot_recovery.max_rate as uint64")
	}

	// Get the slowest range moving rate
	minMoveSpeed := recoveryBytes
	if minMoveSpeed > rebalanceBytes {
		minMoveSpeed = rebalanceBytes
	}

	if len(zones) == 0 {
		zones, err = ZoneConfigs(ctx, db)
		if err != nil {
			return 0, errors.Wrap(err, "failed to retrieve zone configs")
		}
	}

	// Find the largest possible range size
	var maxRangeSize uint64
	for _, zone := range zones {
		if zone.Config.RangeMaxBytes > maxRangeSize {
			maxRangeSize = zone.Config.RangeMaxBytes
		}
	}

	if maxRangeSize == 0 {
		return 0, errors.New("no maximum range size found")
	}

	// Calculate the kindest (values wise, not respecting cluster load) possible duration
	// that it should take for a range to move from one node to another
	return time.Duration(maxRangeSize/minMoveSpeed) * time.Second, nil
}
