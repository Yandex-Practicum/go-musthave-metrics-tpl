package storage

import (
	"sync"

	"github.com/vova4o/yandexadv/internal/models"
)

// MemStorage структура для хранилища в памяти
type MemStorage struct {
	MemStorage map[string]models.Metrics
	mu         sync.Mutex
}

// NewMemStorage создание нового хранилища в памяти
func NewMemStorage() *MemStorage {
	return &MemStorage{
		MemStorage: make(map[string]models.Metrics),
	}
}

// UpdateBatch обновление метрик по пакетно
func (s *MemStorage) UpdateBatch(metrics []models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range metrics {
		s.MemStorage[metric.ID] = metric
	}

	return nil
}

// MetrixStatistic получение статистики метрик
func (s *MemStorage) MetrixStatistic() (map[string]models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var metrics = make(map[string]models.Metrics)

	for metricType, metricValues := range s.MemStorage {
		metrics[metricType] = metricValues
	}

	return metrics, nil
}

// UpdateMetric обновление метрики
func (s *MemStorage) UpdateMetric(metric models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.MemStorage[metric.ID] = metric

	return nil
}

// GetValue получение значения метрики по ID метрики
func (s *MemStorage) GetValue(metric models.Metrics) (*models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if val, ok := s.MemStorage[metric.ID]; ok {
		return &val, nil
	}

	return nil, models.ErrMetricNotFound
}

// Ping проверка подключения к памяти
func (s *MemStorage) Ping() error {
	return nil
}

// Stop завершение работы с хранилищем в памяти
func (s *MemStorage) Stop() error {
	return nil
}
