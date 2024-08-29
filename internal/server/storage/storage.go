package storage

import (
	"sync"

	"github.com/vova4o/yandexadv/internal/models"
)

// Storage структура для хранилища
type Storage struct {
	// Logger     Loggerer
	MemStorage map[string]models.Metrics
	mu         sync.Mutex
}

// // Loggerer интерфейс для логгера
// type Loggerer interface {
// 	Info(string, ...zap.Field)
// 	Error(string, ...zap.Field)
// }

// New создание нового хранилища
func New() *Storage {
	return &Storage{
		MemStorage: make(map[string]models.Metrics),
	}
}

// MetrixStatistic получение статистики метрик
func (s *Storage) MetrixStatistic() (map[string]models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var metrics = make(map[string]models.Metrics)

	for metricType, metricValues := range s.MemStorage {
		metrics[metricType] = metricValues
	}

	return metrics, nil
}

// UpdateMetric обновление метрики
func (s *Storage) UpdateMetric(metric models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.MemStorage[metric.ID] = metric
	return nil
}

// GetValue получение значения метрики по ID метрики
// возвращает значение метрики и ошибку
// возвращает значение не указателем, а значением
func (s *Storage) GetValue(metric models.Metrics) (*models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if val, ok := s.MemStorage[metric.ID]; ok {
		return &val, nil
	}

	return nil, models.ErrMetricNotFound
}
