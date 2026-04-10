package storage

import (
	"context"
	"database/sql"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorege(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) UpdateGauge(ctx context.Context, name string, value float64) {
	s.db.ExecContext(ctx,
		`INSERT INTO metrics(id,type,value) VALUES($1,'gauge', $2)
	ON CONFLICT(id,type) DO UPDATE SET value = $2`, name, value)
}

func (s *PostgresStorage) UpdateCounter(ctx context.Context, name string, delta int64) {
	s.db.ExecContext(ctx,
		`INSERT INTO mertics (id,type,delta) VALUES ($1, 'counter', $2)
		ON CONFLICT (id, type) DO UPDATE SET delta = metrics.delta + $2`, name, delta)
}

func (s *PostgresStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	var value float64
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM metrics WHERE id = $1 AND type = 'gauge'`, name).Scan(&value)
	if err != nil {
		return 0, false
	}
	return value, true
}

func (s *PostgresStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	var delta int64
	err := s.db.QueryRowContext(ctx,
		`SELECT delta FROM metrics WHERE id = $1 AND type = 'counter'`, name).Scan(&delta)
	if err != nil {
		return 0, false
	}
	return delta, true
}

func (s *PostgresStorage) GetAllGauges(ctx context.Context) map[string]float64 {
	result := make(map[string]float64)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, value FROM metrics WHERE type = 'gauge'`)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		rows.Scan(&name, &value)
		result[name] = value
	}
	if err := rows.Err(); err != nil {
		return result
	}
	return result
}

func (s *PostgresStorage) GetAllCounters(ctx context.Context) map[string]int64 {
	result := make(map[string]int64)

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, delta FROM metrics WHERE type = 'counter'`)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var delta int64
		rows.Scan(&name, &delta)
		result[name] = delta
	}
	if err := rows.Err(); err != nil {
		return result
	}
	return result
}
