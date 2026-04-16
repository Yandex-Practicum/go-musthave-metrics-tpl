package storage

import (
	"context"
	"database/sql"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/retry"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = retry.Do(context.Background(), func() error {
		return db.Ping()
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
