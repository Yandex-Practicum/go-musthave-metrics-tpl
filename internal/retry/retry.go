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
		return pgerrcode.IsConnectionException(pgErr.Code)
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
