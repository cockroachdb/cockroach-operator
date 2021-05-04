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

package healthchecker

import (
	"context"
	"database/sql"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

type HealthChecker interface { // for testing
	Probe(ctx context.Context, l logr.Logger, logSuffix string) error
}

type HealthCheckerImpl struct {
	Db *sql.DB
}

func NewHealthChecker(db *sql.DB) *HealthCheckerImpl {
	return &HealthCheckerImpl{Db: db}
}

func (s *HealthCheckerImpl) Probe(ctx context.Context, l logr.Logger, logSuffix string) error {
	l.V(int(zapcore.DebugLevel)).Info("Health check probe", "label", logSuffix)
	return s.waitUntilAllReplicasAreRunning(ctx, l, logSuffix)
}

func (s *HealthCheckerImpl) checkAllReplicasAreRunning(ctx context.Context, l logr.Logger, logSuffix string) error {
	l.V(int(zapcore.DebugLevel)).Info("checkAllReplicasAreRunning", "label", logSuffix)
	underReplicasRangeSum, err := getUnderReplicatedRanges(ctx, s.Db, logSuffix)
	if err != nil {
		return err
	}
	if underReplicasRangeSum != 0 {
		return errors.New("under replica is not zero")
	}
	return nil
}
func (s *HealthCheckerImpl) waitUntilAllReplicasAreRunning(ctx context.Context, l logr.Logger, logSuffix string) error {
	f := func() error {
		return s.checkAllReplicasAreRunning(ctx, l, logSuffix)
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Minute
	b.MaxInterval = 5 * time.Second
	if err := backoff.Retry(f, b); err != nil {
		return errors.Wrapf(err, "replicas check probe failed for cluster %s", logSuffix)
	}
	return nil
}

func getUnderReplicatedRanges(ctx context.Context, db *sql.DB, logSuffix string) (int, error) {
	r := db.QueryRowContext(ctx, "SELECT SUM(under_replicated_ranges) FROM system.replication_stats")
	var value int
	if err := r.Scan(&value); err != nil {
		return -1, errors.Wrapf(err, "failed to get under replicated_ranges for  %s", logSuffix)
	}
	return value, nil
}
