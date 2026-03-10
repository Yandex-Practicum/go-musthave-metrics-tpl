package storage

type Repository interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, delta int64)
	GetGauge(name string) (value float64, ok bool)
	GetCounter(name string) (value int64, ok bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}
