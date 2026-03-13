package storage

import (
	"errors"
	"sync"
)

var (
	ErrMetricNotFound = errors.New("metric not found")
	ErrInvalidType    = errors.New("invalid metric type")
)

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllMetrics() (map[string]float64, map[string]int64)
}

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}

func (m *MemStorage) GetGauge(name string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.gauges[name]
	if !exists {
		return 0, ErrMetricNotFound
	}
	return value, nil
}

func (m *MemStorage) GetCounter(name string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.counters[name]
	if !exists {
		return 0, ErrMetricNotFound
	}
	return value, nil
}

func (m *MemStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gaugesCopy := make(map[string]float64)
	countersCopy := make(map[string]int64)

	for k, v := range m.gauges {
		gaugesCopy[k] = v
	}

	for k, v := range m.counters {
		countersCopy[k] = v
	}

	return gaugesCopy, countersCopy
}
