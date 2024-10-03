package storage

import "sync"

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) SetGauge(metricName string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[metricName] = value
}

func (m *MemStorage) IncrementCounter(metricName string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[metricName] += value
}

func (m *MemStorage) GetGauge(metricName string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, exists := m.gauges[metricName]
	return value, exists
}

func (m *MemStorage) GetCounter(metricName string) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, exists := m.counters[metricName]
	return value, exists
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.gauges
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counters
}
