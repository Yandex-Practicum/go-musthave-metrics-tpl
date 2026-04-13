package service

import (
	"context"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
)

type MetricsService interface {
	UpdateGauge(ctx context.Context, name string, value float64)
	UpdateCounter(ctx context.Context, name string, delta int64)
	GetGauge(ctx context.Context, name string) (float64, bool)
	GetCounter(ctx context.Context, name string) (int64, bool)
	GetAllGauges(ctx context.Context) map[string]float64
	GetAllCounters(ctx context.Context) map[string]int64
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
}
