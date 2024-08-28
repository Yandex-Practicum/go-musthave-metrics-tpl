package storage

import (
	"sync"

	"github.com/vova4o/yandexadv/internal/models"
)

// Storage структура для хранилища
type Storage struct {
	// Logger     Loggerer
	MemStorage map[string]models.Metric
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
		MemStorage: make(map[string]models.Metric),
	}
}

// MetrixStatistic получение статистики метрик
func (s *Storage) MetrixStatistic() (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var metrics = make(map[string]interface{})

	for metricType, metricValues := range s.MemStorage {
		metrics[metricType] = metricValues
	}

	return metrics, nil
}

// UpdateMetric обновление метрики
func (s *Storage) UpdateMetric(metric models.Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.MemStorage[metric.Name] = metric
	return nil
}

// GetValue получение значения метрики
func (s *Storage) GetValue(metric models.Metric) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if value, ok := s.MemStorage[metric.Name]; ok {
		return value.Value, nil
	}

	return "", models.ErrMetricNotFound
}
