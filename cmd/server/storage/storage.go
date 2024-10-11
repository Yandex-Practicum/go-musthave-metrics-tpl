package storage

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MemStorage) SetGauge(metricName string, value float64) {
	m.gauges[metricName] = value
}

func (m *MemStorage) IncrementCounter(metricName string, value int64) {
	m.counters[metricName] += value
}

func (m *MemStorage) GetGauge(metricName string) (float64, bool) {
	value, exists := m.gauges[metricName]
	return value, exists
}

func (m *MemStorage) GetCounter(metricName string) (int64, bool) {
	value, exists := m.counters[metricName]
	return value, exists
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	return m.gauges
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	return m.counters
}
