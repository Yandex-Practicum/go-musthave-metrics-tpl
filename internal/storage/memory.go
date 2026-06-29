package storage

import (
	"context"
	"sync"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
)

type MemoryStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
	syncFile string
}

func (s *MemoryStorage) SetSyncFile(path string) {
	s.syncFile = path
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemoryStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	s.mu.Lock()
	s.gauges[name] = value
	s.mu.Unlock()
	if s.syncFile != "" {
		SaveToFile(s, s.syncFile)
	}
	return nil
}

func (s *MemoryStorage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	s.mu.Lock()
	s.counters[name] += delta
	s.mu.Unlock()
	if s.syncFile != "" {
		SaveToFile(s, s.syncFile)
	}
	return nil
}

func (s *MemoryStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.gauges[name]
	return v, ok
}

func (s *MemoryStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.counters[name]
	return v, ok
}

func (s *MemoryStorage) GetAllGauges(ctx context.Context) map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		result[k] = v
	}
	return result
}

func (s *MemoryStorage) GetAllCounters(ctx context.Context) map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		result[k] = v
	}
	return result
}

func (s *MemoryStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	s.mu.Lock()
	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			if m.Value != nil {
				s.gauges[m.ID] = *m.Value
			}
		case "counter":
			if m.Delta != nil {
				s.counters[m.ID] += *m.Delta
			}
		}
	}
	s.mu.Unlock()
	if s.syncFile != "" {
		return SaveToFile(s, s.syncFile)
	}
	return nil
}

// snapshot собирает все метрики в один срез под одной блокировкой чтения.
// Точная ёмкость backing-массивов исключает реаллокацию, поэтому взятие
// адресов их элементов безопасно.
func (s *MemoryStorage) snapshot() []models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(s.gauges)+len(s.counters))
	gv := make([]float64, 0, len(s.gauges))
	for name, value := range s.gauges {
		gv = append(gv, value)
		metrics = append(metrics, models.Metrics{ID: name, MType: "gauge", Value: &gv[len(gv)-1]})
	}
	cv := make([]int64, 0, len(s.counters))
	for name, value := range s.counters {
		cv = append(cv, value)
		metrics = append(metrics, models.Metrics{ID: name, MType: "counter", Delta: &cv[len(cv)-1]})
	}
	return metrics
}
