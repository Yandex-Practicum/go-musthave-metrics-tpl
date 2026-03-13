// Package model provides data structures for metrics.
package model

// MetricType represents the type of a metric
type MetricType string

// Metric types
const (
	TypeCounter MetricType = "counter"
	TypeGauge   MetricType = "gauge"
)

// Metrics represents a metric with its type, value, and other properties.
// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	// ID is the name of the metric
	ID string `json:"id"`
	// MType is the type of the metric, either "gauge" or "counter"
	MType MetricType `json:"type"`
	// Delta is the value of a counter metric
	Delta *int64 `json:"delta,omitempty"`
	// Value is the value of a gauge metric
	Value *float64 `json:"value,omitempty"`
	// Hash is the hash of the metric
	Hash string `json:"hash,omitempty"`
}
