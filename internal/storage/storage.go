// Package storage provides an in-memory storage implementation for metrics.
package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
)

// ErrMetricNotFound is returned when a requested metric is not found.
var (
	ErrMetricNotFound = errors.New("metric not found")
	ErrInvalidType    = errors.New("invalid metric type")
)

// Storage interface defines the contract for metric storage operations.
type Storage interface {
	// UpdateGauge updates a gauge metric with the given name and value.
	UpdateGauge(name string, value float64)
	// UpdateCounter updates a counter metric with the given name and value.
	UpdateCounter(name string, value int64)
	// GetGauge retrieves the value of a gauge metric by name.
	GetGauge(name string) (float64, error)
	// GetCounter retrieves the value of a counter metric by name.
	GetCounter(name string) (int64, error)
	// GetAllMetrics returns all stored metrics.
	GetAllMetrics() (map[string]float64, map[string]int64)
	Ping(ctx context.Context) error
	Close() error
}

// MemStorage is an in-memory implementation of the Storage interface.
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

// NewMemStorage creates and returns a new MemStorage instance.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// UpdateGauge updates a gauge metric with the given name and value.
// It is safe for concurrent use.
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

// UpdateCounter updates a counter metric with the given name and value.
// It is safe for concurrent use.
func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}

// GetGauge retrieves the value of a gauge metric by name.
// Returns an error if the metric is not found.
// It is safe for concurrent use.
func (m *MemStorage) GetGauge(name string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.gauges[name]
	if !exists {
		return 0, ErrMetricNotFound
	}
	return value, nil
}

// GetCounter retrieves the value of a counter metric by name.
// Returns an error if the metric is not found.
// It is safe for concurrent use.
func (m *MemStorage) GetCounter(name string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.counters[name]
	if !exists {
		return 0, ErrMetricNotFound
	}
	return value, nil
}

// GetAllMetrics returns all stored metrics.
// It is safe for concurrent use.
func (m *MemStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gaugesCopy := make(map[string]float64, len(m.gauges))
	countersCopy := make(map[string]int64, len(m.counters))

	for k, v := range m.gauges {
		gaugesCopy[k] = v
	}

	for k, v := range m.counters {
		countersCopy[k] = v
	}

	return gaugesCopy, countersCopy
}

func (m *MetricsStorage) Ping(ctx context.Context) error {
	// Для in-memory хранилища всегда возвращаем успешный ping
	return nil
}
