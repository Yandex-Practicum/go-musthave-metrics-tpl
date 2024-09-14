package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vova4o/yandexadv/internal/models"
	"github.com/vova4o/yandexadv/internal/server/flags"
	"go.uber.org/zap"
)

// DBStorage структура для хранилища
type DBStorage struct {
	DB *pgxpool.Pool
	logger Loggerer
}

const maxRetries = 3
const retryDelay = 1 * time.Second

// DBConnect подключение к базе данных
func DBConnect(config *flags.Config, logger Loggerer) (*DBStorage, error) {
	var db *pgxpool.Pool
	var err error

	for i := 0; i < maxRetries; i++ {
        db, err = pgxpool.Connect(context.Background(), config.DBDSN)
        if err == nil {
            break
        }

        logger.Error("Failed to connect to database", zap.Error(err))
        time.Sleep(retryDelay)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
    }

	return &DBStorage{
		DB: db,
		logger: logger,
		}, nil
}

// Ping проверка подключения к базе данных
func (d *DBStorage) Ping() error {
	if d.DB == nil {
		return fmt.Errorf("database is not connected")
	}
	return d.DB.Ping(context.Background())
}

// Stop закрытие подключения к базе данных
func (d *DBStorage) Stop() error {
	if d.DB == nil {
		return nil
	}
	d.DB.Close()
	return nil
}

// CreateTables создание таблиц
func (d *DBStorage) CreateTables() error {
	_, err := d.DB.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS metrics (
		id SERIAL PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL UNIQUE,
		value DOUBLE PRECISION,
		delta BIGINT,
		timestamp TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics (name);`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// UpdateBatch обновление метрик
func (d *DBStorage) UpdateBatch(metrics []models.Metrics) error {
	d.logger.Info("UpdateBatch", zap.String("metrics", fmt.Sprintf("%v", metrics)))

	copyCount, err := d.DB.CopyFrom(
		context.Background(),
		pgx.Identifier{"metrics"},
		[]string{"name", "type", "value", "delta", "timestamp"},
		pgx.CopyFromSlice(len(metrics), func(i int) ([]interface{}, error) {
			return []interface{}{
				metrics[i].ID,
				metrics[i].MType,
				metrics[i].Value,
				metrics[i].Delta,
				time.Now(),
			}, nil
		}),
	)
	if err != nil {
		log.Println("Db failed to insert", err)
		return fmt.Errorf("failed to copy data: %w", err)
	}

	log.Printf("Inserted %d rows", copyCount)

	return nil
}

// UpdateMetric добавление метрики
func (d *DBStorage) UpdateMetric(metric models.Metrics) error {
	_, err := d.DB.Exec(context.Background(), `INSERT INTO metrics (type, name, value, delta, timestamp)
	VALUES ($1, $2, $3, $4, $5)
	 ON CONFLICT (name) DO UPDATE SET
        type = EXCLUDED.type,
        value = EXCLUDED.value,
        delta = EXCLUDED.delta,
        timestamp = EXCLUDED.timestamp`,
		metric.MType, metric.ID, metric.Value, metric.Delta, time.Now())
	if err != nil {
		log.Println("Db failed to insert", err)
		return fmt.Errorf("failed to insert metric: %w", err)
	}
	return nil
}

// // UpdateMetric добавление метрики
// func (d *DBStorage) UpdateMetric(metric models.Metrics) error {
// 	_, err := d.DB.Exec(context.Background(), `INSERT INTO metrics (type, name, value, delta, timestamp)
// 	VALUES ($1, $2, $3, $4, $5)`,
// 		metric.MType, metric.ID, metric.Value, metric.Delta, time.Now())
// 	if err != nil {
// 		log.Println("Db failed to insert", err)
// 		return fmt.Errorf("failed to insert metric: %w", err)
// 	}
// 	return nil
// }

// MetrixStatistic получение статистики метрик
func (d *DBStorage) MetrixStatistic() (map[string]models.Metrics, error) {
    query := `
        SELECT id, type, name, value, delta, timestamp
        FROM (
            SELECT id, type, name, value, delta, timestamp,
                ROW_NUMBER() OVER (PARTITION BY name ORDER BY timestamp DESC) as rn
            FROM metrics
        ) subquery
        WHERE rn = 1;
    `

    rows, err := d.DB.Query(context.Background(), query)
    if err != nil {
        return nil, fmt.Errorf("failed to select metrics: %w", err)
    }
    defer rows.Close()

    metrics := make(map[string]models.Metrics)
    for rows.Next() {
        var metric models.Metrics
        var id int
        var timestamp time.Time
        err = rows.Scan(&id, &metric.MType, &metric.ID, &metric.Value, &metric.Delta, &timestamp)
        if err != nil {
            return nil, fmt.Errorf("failed to scan metrics: %w", err)
        }
        metrics[metric.ID] = metric
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("failed to iterate over metrics: %w", err)
    }

    return metrics, nil
}

// GetValue получение значения метрики по ID метрики
func (d *DBStorage) GetValue(metric models.Metrics) (*models.Metrics, error) {
    row := d.DB.QueryRow(context.Background(), `SELECT id, type, name, value, delta, timestamp FROM metrics WHERE name = $1 ORDER BY timestamp DESC LIMIT 1`, metric.ID)

    var m models.Metrics
    var id int
    var timestamp time.Time
    err := row.Scan(&id, &m.MType, &m.ID, &m.Value, &m.Delta, &timestamp)
    if err != nil {
        if err == pgx.ErrNoRows {
            // Если метрика не найдена, возвращаем значение по умолчанию
            m.Value = nil
            m.Delta = nil
            return &m, nil
        }
        return nil, fmt.Errorf("failed to select metric: %w", err)
    }

    return &m, nil
}