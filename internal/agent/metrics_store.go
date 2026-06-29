package agent

import "sync"

type MetricsStore struct {
	mu        sync.Mutex
	gauges    []GaugeMetric
	psMetrics []GaugeMetric
	pollCount int64
}

func NewMetricsStore() *MetricsStore {
	return &MetricsStore{}
}

func (s *MetricsStore) UpdateGauges(gauges []GaugeMetric) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges = gauges
	s.pollCount++
}

func (s *MetricsStore) UpdatePSMetrics(metrics []GaugeMetric) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.psMetrics = metrics
}

func (s *MetricsStore) GetAll() ([]GaugeMetric, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := append(s.gauges, s.psMetrics...)
	return all, s.pollCount
}
