package server

import (
	"context"
	"yandex-go-musthave-metrics-tpl/internal/model"
	storage "yandex-go-musthave-metrics-tpl/internal/repository/memstorage"
)

//go:generate mockgen -package server -source=metrics.go -destination=mock_metrics_test.go
var _ MetricRepository = (*storage.MemStorage)(nil)

type MetricRepository interface {
	Add(ctx context.Context, m model.Metrics) error
	GetValue(ctx context.Context, name string) (*float64, error)
	GetAll(ctx context.Context) ([]model.Metrics, error)
}

type MetricService struct {
	repository MetricRepository
}

func NewMetricService(mr MetricRepository) *MetricService {
	return &MetricService{
		repository: mr,
	}
}

func (s *MetricService) Add(ctx context.Context, m model.Metrics) error {
	if m.MType == model.Counter {
		oldMetricVal, err := s.repository.GetValue(ctx, m.ID)
		if err == nil {
			*m.Value += *oldMetricVal
		}
	}
	if err := s.repository.Add(ctx, m); err != nil {
		return err
	}
	return nil
}

func (s *MetricService) GetValue(ctx context.Context, name string) (*float64, error) {
	metricValue, err := s.repository.GetValue(ctx, name)
	if err != nil {
		return nil, err
	}
	return metricValue, nil
}

func (s *MetricService) GetAll(ctx context.Context) ([]model.Metrics, error) {
	metrics, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}
