package service

import (
	"context"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

type metricService struct {
	repo storage.Repository
}

func NewMetricsService(repo storage.Repository) MetricsService {
	return &metricService{repo: repo}
}

func (s *metricService) UpdateGauge(ctx context.Context, name string, value float64) {
	s.repo.UpdateGauge(ctx, name, value)
}

func (s *metricService) UpdateCounter(ctx context.Context, name string, delta int64) {
	s.repo.UpdateCounter(ctx, name, delta)
}

func (s *metricService) GetGauge(ctx context.Context, name string) (float64, bool) {
	return s.repo.GetGauge(ctx, name)
}

func (s *metricService) GetCounter(ctx context.Context, name string) (int64, bool) {
	return s.repo.GetCounter(ctx, name)
}

func (s *metricService) GetAllGauges(ctx context.Context) map[string]float64 {
	return s.repo.GetAllGauges(ctx)
}

func (s *metricService) GetAllCounters(ctx context.Context) map[string]int64 {
	return s.repo.GetAllCounters(ctx)
}

func (s *metricService) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	return s.repo.UpdateBatch(ctx, metrics)
}
