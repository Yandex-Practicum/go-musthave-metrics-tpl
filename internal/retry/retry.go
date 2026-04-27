package retry

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var intervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch {
		case pgerrcode.IsConnectionException(pgErr.Code): //Class 08
			return true
		case pgErr.Code == "40001": // serialization_failure
			return true
		case pgErr.Code == "40P01": // deadlock_detected
			return true
		case pgErr.Code == "57P01": // admin_shutdown
			return true
		case pgErr.Code == "53300": // too_many_connections
			return true
		default:
			return false
		}
	}
	return true
}

func Do(ctx context.Context, fn func() error) error {
	err := fn()
	if err == nil || !IsRetriable(err) {
		return err
	}

	for _, interval := range intervals {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		err = fn()
		if err == nil || !IsRetriable(err) {
			return err
		}
	}
	return err
}
