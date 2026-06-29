package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPSink — приёмник аудита, отправляющий события методом POST
// на указанный удалённый URL.
type HTTPSink struct {
	url    string
	client *http.Client
}

// NewHTTPSink создаёт HTTP-приёмник аудита.
func NewHTTPSink(url string) *HTTPSink {
	return &HTTPSink{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Notify отправляет событие методом POST в формате JSON.
func (s *HTTPSink) Notify(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("приёмник аудита вернул статус %d", resp.StatusCode)
	}
	return nil
}
