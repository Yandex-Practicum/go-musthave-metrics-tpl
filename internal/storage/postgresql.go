package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func (p *PostgresStorage) Close() {
	panic("unimplemented")
}

func NewPostgresStorage(ctx context.Context, dsn string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	return &PostgresStorage{pool: pool}, nil
}

func (p *PostgresStorage) BatchUpdate(ctx context.Context, metrics []model.Metrics) error {
	// Используем pool для начала транзакции
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction begin error: %w", err)
	}
	defer tx.Rollback(ctx) // Добавляем контекст

	const (
		counterQuery = `INSERT INTO counters (name, value)
                      VALUES ($1, $2)
                      ON CONFLICT (name) DO UPDATE
                      SET value = counters.value + EXCLUDED.value`

		gaugeQuery = `INSERT INTO gauges (name, value)
                    VALUES ($1, $2)
                    ON CONFLICT (name) DO UPDATE
                    SET value = EXCLUDED.value`
	)

	for _, m := range metrics {
		switch m.MType {
		case model.TypeCounter:
			if m.Delta == nil {
				continue // Пропускаем некорректные метрики
			}
			if _, err = tx.Exec(ctx, counterQuery, m.ID, *m.Delta); err != nil {
				return fmt.Errorf("counter update failed: %w", err)
			}

		case model.TypeGauge:
			if m.Value == nil {
				continue // Пропускаем некорректные метрики
			}
			if _, err = tx.Exec(ctx, gaugeQuery, m.ID, *m.Value); err != nil {
				return fmt.Errorf("gauge update failed: %w", err)
			}

		default:
			return fmt.Errorf("unknown metric type: %s", m.MType)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction commit error: %w", err)
	}
	return nil
}

// Пример реализации метода для счетчиков
func (p *PostgresStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO counters (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE
		SET value = counters.value + EXCLUDED.value`,
		name, value)
	return err
}

func (p *PostgresStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO gauges (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE
		SET value = EXCLUDED.value`,
		name, value)
	return err
}

func (p *PostgresStorage) GetAllMetrics(ctx context.Context) (map[string]float64, map[string]int64, error) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	// Получение gauge метрик
	rows, err := p.pool.Query(ctx, "SELECT name, value FROM gauges")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, nil, err
		}
		gauges[name] = value
	}

	// Получение counter метрик
	rows, err = p.pool.Query(ctx, "SELECT name, value FROM counters")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, nil, err
		}
		counters[name] = value
	}

	return gauges, counters, nil
}
