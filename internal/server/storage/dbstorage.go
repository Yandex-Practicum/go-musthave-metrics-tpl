package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	// driver to connect to PostgreSQL
	_ "github.com/lib/pq"
	"github.com/vova4o/yandexadv/internal/models"
	"github.com/vova4o/yandexadv/internal/server/flags"
)

// DBStorage структура для хранилища
type DBStorage struct {
	DB *sql.DB
}

// DBConnect подключение к базе данных
func DBConnect(config *flags.Config) (*DBStorage, error) {
	db, err := sql.Open("postgres", config.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBStorage{DB: db}, nil
}

// Ping проверка подключения к базе данных
func (d *DBStorage) Ping() error {
	if d.DB == nil {
		return fmt.Errorf("database is not connected")
	}
	return d.DB.Ping()
}

// Stop закрытие подключения к базе данных
func (d *DBStorage) Stop() error {
	if d.DB == nil {
		return nil
	}
	return d.DB.Close()
}

// CreateTables создание таблиц
func (d *DBStorage) CreateTables() error {
	_, err := d.DB.Exec(`CREATE TABLE IF NOT EXISTS metrics (
		id SERIAL PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
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

// UpdateMetric добавление метрики
func (d *DBStorage) UpdateMetric(metric models.Metrics) error {

	_, err := d.DB.Exec(`INSERT INTO metrics (type, name, value, delta, timestamp)
	VALUES ($1, $2, $3, $4, $5)`,
		metric.MType, metric.ID, metric.Value, metric.Delta, time.Now())
	if err != nil {
		log.Println("Db faild to insert",err)
		return fmt.Errorf("failed to insert metric: %w", err)
	}
	return nil
}

// MetrixStatistic получение статистики метрик
func (d *DBStorage) MetrixStatistic() (map[string]models.Metrics, error) {
	rows, err := d.DB.Query(`SELECT * FROM metrics`)
	if err != nil {
		return nil, fmt.Errorf("failed to select metrics: %w", err)
	}
	defer rows.Close()

	//check if there are no metrics
	if !rows.Next() {
		return nil, fmt.Errorf("no metrics found")
	}

	var metrics = make(map[string]models.Metrics)
	for rows.Next() {
		var metric models.Metrics
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
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
    row := d.DB.QueryRow(`SELECT id, type, name, value, delta, timestamp FROM metrics WHERE name = $1`, metric.ID)

    var m models.Metrics
    var id int
    var timestamp sql.NullTime
    err := row.Scan(&id, &m.MType, &m.ID, &m.Value, &m.Delta, &timestamp)
    if err != nil {
        if err == sql.ErrNoRows {
            // Если метрика не найдена, возвращаем значение по умолчанию
            m.Value = nil
            m.Delta = nil
            return &m, nil
        }
        return nil, fmt.Errorf("failed to select metric: %w", err)
    }

    return &m, nil
}