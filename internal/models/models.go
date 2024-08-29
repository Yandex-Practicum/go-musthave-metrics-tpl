package models

import "errors"

// Metric структура для метрик
type Metric struct {
	Type  string      `json:"type"`
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// Metrics структура для метрик с типом и значением
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// HTTPError структура для ошибок с HTTP-статусом
type HTTPError struct {
	Status  int
	Message string
}

// MetricsError готовые ошибки
var (
	ErrMetricTypeNotFound = errors.New("metric type not found")
	ErrMetricNotFound     = errors.New("metric not found")
)

// Error реализация интерфейса ошибки
func (e *HTTPError) Error() string {
	return e.Message
}

// NewHTTPError создание новой ошибки с HTTP-статусом
func NewHTTPError(status int, message string) *HTTPError {
	return &HTTPError{
		Status:  status,
		Message: message,
	}
}
