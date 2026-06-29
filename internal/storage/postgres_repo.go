package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/retry"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorege(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	return retry.Do(ctx, func() error {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO metrics(id,type,value) VALUES($1,'gauge', $2)
	ON CONFLICT(id,type) DO UPDATE SET value = $2`, name, value)
		return err
	})
}

func (s *PostgresStorage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	return retry.Do(ctx, func() error {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO metrics (id,type,delta) VALUES ($1, 'counter', $2)
		ON CONFLICT (id, type) DO UPDATE SET delta = metrics.delta + $2`, name, delta)
		return err
	})
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
		return make(map[string]float64)
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
		return make(map[string]int64)
	}
	return result
}

func (s *PostgresStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	return retry.Do(ctx, func() error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		gaugeStmt, err := tx.PrepareContext(ctx,
			`INSERT INTO metrics (id,type,value) VALUES ($1, 'gauge', $2)
		ON CONFLICT (id, type) DO UPDATE SET value =$2`)
		if err != nil {
			return err
		}
		defer gaugeStmt.Close()

		var gaugeArgs []interface{}
		var counterArgs []interface{}
		var gaugeValues []string
		var counterValues []string
		gIdx := 1
		cIdx := 1

		for _, m := range metrics {
			switch m.MType {
			case "gauge":
				if m.Value != nil {
					gaugeValues = append(gaugeValues, fmt.Sprintf("($%d, 'gauge', $%d)", gIdx, gIdx+1))
					gaugeArgs = append(gaugeArgs, m.ID, *m.Value)
					gIdx += 2
				}
			case "counter":
				if m.Delta != nil {
					counterValues = append(counterValues, fmt.Sprintf("($%d, 'counter', $%d)", cIdx, cIdx+1))
					counterArgs = append(counterArgs, m.ID, *m.Delta)
					cIdx += 2
				}
			}
		}

		if len(gaugeValues) > 0 {
			query := fmt.Sprintf(
				`INSERT INTO metrics (id, type, value) VALUES %s
			ON CONFLICT (id, type) DO UPDATE SET value = EXCLUDED.value`,
				strings.Join(gaugeValues, ","))
			if _, err := tx.ExecContext(ctx, query, gaugeArgs...); err != nil {
				return err
			}
		}

		if len(counterValues) > 0 {
			query := fmt.Sprintf(
				`INSERT INTO metrics (id, type, value) VALUES %s
			ON CONFLICT (id, type) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta`,
				strings.Join(counterValues, ","))
			if _, err := tx.ExecContext(ctx, query, counterArgs...); err != nil {
				return err
			}
		}

		return tx.Commit()
	})
}
