/*
Copyright 2023 The Cockroach Authors

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

package clustersql_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/cockroachdb/cockroach-operator/pkg/clustersql"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func TestGetClusterSetting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	t.Run("return setting value when found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"name"}).AddRow("bogus_value")
		mock.ExpectQuery("SHOW CLUSTER SETTING bogus_setting").WillReturnRows(rows)

		v, err := GetClusterSetting(context.Background(), db, "bogus_setting")
		require.Equal(t, "bogus_value", v)
		require.NoError(t, err)
	})

	t.Run("returns error with an invalid setting name", func(t *testing.T) {
		v, err := GetClusterSetting(context.Background(), db, "!not#valid$")
		require.Empty(t, v)
		require.Equal(t, ErrInvalidClusterSettingName, errors.Cause(err))
	})

	t.Run("returns error when setting not found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"name"})
		mock.ExpectQuery("SHOW CLUSTER SETTING bogus_setting").WillReturnRows(rows)

		v, err := GetClusterSetting(context.Background(), db, "bogus_setting")
		require.Empty(t, v)
		require.Equal(t, sql.ErrNoRows, errors.Cause(err))
	})
}

func TestSetClusterSetting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	t.Run("returns no error when setting value", func(t *testing.T) {
		mock.
			ExpectExec("SET CLUSTER SETTING bogus_setting = \\$1").
			WithArgs("bogus_value").
			WillReturnResult(sqlmock.NewResult(1, 1))

		require.NoError(t, SetClusterSetting(context.Background(), db, "bogus_setting", "bogus_value"))
	})

	t.Run("returns error with an invalid setting name", func(t *testing.T) {
		err := SetClusterSetting(context.Background(), db, "!not#valid$", "nope")
		require.Equal(t, ErrInvalidClusterSettingName, errors.Cause(err))
	})

	t.Run("returns error when exec fails", func(t *testing.T) {
		mock.
			ExpectExec("SET CLUSTER SETTING bogus_setting = \\$1").
			WithArgs("nope").
			WillReturnError(errors.New("boom"))

		err := SetClusterSetting(context.Background(), db, "bogus_setting", "nope")
		require.EqualError(t, errors.Cause(err), "boom")
	})
}

func TestRangeMoveDuration(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	stubSettings := func(settings map[string]string) {
		// map keys are not ordered
		mock.MatchExpectationsInOrder(false)

		for k, v := range settings {
			mock.
				ExpectQuery("SHOW CLUSTER SETTING " + k).
				WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(v))
		}
	}

	t.Run("calculates the time it takes for a range to move nodes", func(t *testing.T) {
		tests := []struct {
			name             string
			maxRebalanceRate string
			maxRecoveryRate  string
			zones            []Zone
			exp              time.Duration
		}{
			{
				name:             "rebalance rate faster than recovery",
				maxRebalanceRate: "10MB",
				maxRecoveryRate:  "6MB",
				zones: []Zone{
					{Target: "zone-1", Config: ZoneConfig{RangeMaxBytes: 10}},
					{Target: "zone-2", Config: ZoneConfig{RangeMaxBytes: 20}},
				},
				// TODO: Is it required. Output will always be 0s because of integer division
				// nolint
				exp: time.Duration(10/1000000) * time.Second,
			},
			{
				name:             "recovery rate faster than rebalance",
				maxRebalanceRate: "1MB",
				maxRecoveryRate:  "2MB",
				zones: []Zone{
					{Target: "zone-1", Config: ZoneConfig{RangeMaxBytes: 10}},
					{Target: "zone-2", Config: ZoneConfig{RangeMaxBytes: 20}},
				},
				// TODO: Is it required. Output will always be 0s because of integer division
				// nolint
				exp: time.Duration(20/1000000) * time.Second,
			},
			{
				name:             "min range size is used",
				maxRebalanceRate: "1MB",
				maxRecoveryRate:  "2MB",
				zones: []Zone{
					{Target: "zone-1", Config: ZoneConfig{RangeMaxBytes: 20}},
					{Target: "zone-2", Config: ZoneConfig{RangeMaxBytes: 30}},
				},
				// TODO: Is it required. Output will always be 0s because of integer division
				// nolint
				exp: time.Duration(20/2000000) * time.Second,
			},
		}

		for _, tt := range tests {
			stubSettings(map[string]string{
				"kv.snapshot_rebalance.max_rate": tt.maxRebalanceRate,
				"kv.snapshot_recovery.max_rate":  tt.maxRecoveryRate,
			})

			d, err := RangeMoveDuration(ctx, db, tt.zones...)
			require.Equal(t, tt.exp, d)
			require.NoError(t, err)
		}
	})

	t.Run("returns error when MaxRangeSize is 0 for all zones", func(t *testing.T) {
		stubSettings(map[string]string{
			"kv.snapshot_rebalance.max_rate": "10MB",
			"kv.snapshot_recovery.max_rate":  "10MB",
		})

		d, err := RangeMoveDuration(ctx, db, Zone{
			Target: "us-east1-c",
			Config: ZoneConfig{RangeMaxBytes: 0},
		})

		require.Zero(t, d)
		require.EqualError(t, err, "no maximum range size found")
	})

	t.Run("return error when fetching ZoneConfigs fails", func(t *testing.T) {
		stubSettings(map[string]string{
			"kv.snapshot_rebalance.max_rate": "10MB",
			"kv.snapshot_recovery.max_rate":  "10MB",
		})

		// TODO (pseudomuto): this feels a bit dirty reaching into SQL used within zones.go
		mock.ExpectQuery("SELECT target, full_config_yaml FROM crdb_internal.zones").WillReturnError(errors.New("boom"))

		d, err := RangeMoveDuration(ctx, db)
		require.Zero(t, d)
		require.EqualError(t, errors.Cause(err), "boom")
	})

	t.Run("returns error when setting not found", func(t *testing.T) {
		// These are order dependent in that they are queried in this exact order in RangeMoveDuration. This is necessary to
		// make it easier to stub previous values when validating them in a loop.
		tests := []string{
			"kv.snapshot_rebalance.max_rate",
			"kv.snapshot_recovery.max_rate",
		}

		for i, tt := range tests {
			for j := 0; j < i; j++ {
				// all previous values should be valid
				mock.
					ExpectQuery("SHOW CLUSTER SETTING " + tests[j]).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("1MB"))
			}

			// this particular one is not found
			mock.ExpectQuery("SHOW CLUSTER SETTING " + tt).WillReturnRows(sqlmock.NewRows([]string{"name"}))

			d, err := RangeMoveDuration(ctx, db)
			require.Zero(t, d)
			require.Contains(t, err.Error(), "failed to get "+tt)
		}
	})

	t.Run("returns error when settings aren't valid", func(t *testing.T) {
		tests := []struct {
			badKey   string
			settings map[string]string
		}{
			{
				badKey: "kv.snapshot_rebalance.max_rate",
				settings: map[string]string{
					"kv.snapshot_rebalance.max_rate": "-1Mb",
					"kv.snapshot_recovery.max_rate":  "1Mb",
				},
			},
			{
				badKey: "kv.snapshot_recovery.max_rate",
				settings: map[string]string{
					"kv.snapshot_rebalance.max_rate": "1Mb",
					"kv.snapshot_recovery.max_rate":  "-1Mb",
				},
			},
		}

		for _, tt := range tests {
			stubSettings(tt.settings)

			d, err := RangeMoveDuration(ctx, db)
			require.Zero(t, d)
			require.Contains(t, err.Error(), fmt.Sprintf("failed to parse %s as uint64", tt.badKey))
		}
	})
}
