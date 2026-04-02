package storage

import (
	"context"
	"sync"
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

func (s *MemoryStorage) UpdateGauge(ctx context.Context, name string, value float64) {
	s.mu.Lock()
	s.gauges[name] = value
	s.mu.Unlock()
	if s.syncFile != "" {
		SaveToFile(s, s.syncFile)
	}
}

func (s *MemoryStorage) UpdateCounter(ctx context.Context, name string, delta int64) {
	s.mu.Lock()
	s.counters[name] += delta
	s.mu.Unlock()
	if s.syncFile != "" {
		SaveToFile(s, s.syncFile)
	}
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
