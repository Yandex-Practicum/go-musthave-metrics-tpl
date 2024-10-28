package memstorage

type MemStorage struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func (m *MemStorage) SetGauge(metricName string, value float64) {
	m.Gauges[metricName] = value
}

func (m *MemStorage) IncrementCounter(metricName string, value int64) {
	m.Counters[metricName] += value
}

func (m *MemStorage) GetGauge(metricName string) (float64, bool) {
	value, exists := m.Gauges[metricName]
	return value, exists
}

func (m *MemStorage) GetCounter(metricName string) (int64, bool) {
	value, exists := m.Counters[metricName]
	return value, exists
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	return m.Gauges
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	return m.Counters
}
