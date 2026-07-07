// Package storage содержит реализации хранилищ метрик: в памяти,
// в файле и в PostgreSQL.
package storage

import (
	"context"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
)

type Repository interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, delta int64) error
	GetGauge(ctx context.Context, name string) (value float64, ok bool)
	GetCounter(ctx context.Context, name string) (value int64, ok bool)
	GetAllGauges(ctx context.Context) map[string]float64
	GetAllCounters(ctx context.Context) map[string]int64
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
}
