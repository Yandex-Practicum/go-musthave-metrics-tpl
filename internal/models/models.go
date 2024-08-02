package models

// HTTPError структура для ошибок с HTTP-статусом
type HTTPError struct {
	Status  int
	Message string
}

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