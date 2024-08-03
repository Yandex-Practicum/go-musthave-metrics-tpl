package models

import "errors"

type Metric struct {
	Type  string      `json:"type"`
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// HTTPError структура для ошибок с HTTP-статусом
type HTTPError struct {
	Status  int
	Message string
}

// MetricsError готовые ошибки
var (
	ErrMetricTypeNotFound = errors.New("metric type not found")
	ErrMetricNotFound =    errors.New("metric not found")
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
