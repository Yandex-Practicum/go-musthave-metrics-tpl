// Package models описывает модель метрик и её JSON-представление,
// используемое при передаче данных между агентом и сервером.
package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics описывает метрику: её идентификатор, тип и значение.
//
// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
