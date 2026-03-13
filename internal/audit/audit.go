package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

// AuditEvent представляет событие аудита
type AuditEvent struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// Auditor интерфейс для наблюдателей аудита
type Auditor interface {
	Notify(event AuditEvent) error
}

// FileAuditor реализует аудит в файл
type FileAuditor struct {
	filePath string
}

// HTTPAuditor реализует аудит через HTTP POST
type HTTPAuditor struct {
	url string
}

// NewFileAuditor создает нового аудитора для файла
func NewFileAuditor(filePath string) *FileAuditor {
	return &FileAuditor{filePath: filePath}
}

// NewHTTPAuditor создает нового аудитора для HTTP
func NewHTTPAuditor(url string) *HTTPAuditor {
	return &HTTPAuditor{url: url}
}

// Notify отправляет событие аудита в файл
func (f *FileAuditor) Notify(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Добавляем перевод строки в конце
	data = append(data, '\n')

	// Открываем файл в режиме добавления
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// Notify отправляет событие аудита по HTTP
func (h *HTTPAuditor) Notify(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = http.Post(h.url, "application/json", bytes.NewReader(data))
	return err
}
