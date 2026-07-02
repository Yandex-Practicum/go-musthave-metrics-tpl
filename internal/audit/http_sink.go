package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Настройки HTTP-приёмника.
const (
	// httpSinkTimeout ограничивает время ожидания всей операции отправки
	// (включая ретраи) через context в Notify.
	httpSinkTimeout = 5 * time.Second
	// httpSinkRetryMax — количество дополнительных попыток при временных
	// сетевых ошибках. Основная попытка не считается.
	httpSinkRetryMax = 3
	// httpSinkRetryWaitMin и httpSinkRetryWaitMax — границы
	// экспоненциального ожидания между попытками.
	httpSinkRetryWaitMin = 200 * time.Millisecond
	httpSinkRetryWaitMax = 2 * time.Second
)

// HTTPSink — приёмник аудита, отправляющий события методом POST
// на указанный удалённый URL. Ретраи выполняются автоматически
// обёрткой retryablehttp вокруг стандартного *http.Client — код
// Notify остаётся прежним.
type HTTPSink struct {
	url    string
	client *http.Client
}

// NewHTTPSink создаёт HTTP-приёмник аудита. HTTP-клиент оборачивается
// в retryablehttp для прозрачных ретраев при временных сетевых ошибках
// (5xx, коннекторные ошибки). Время всей операции ограничивается через
// context в Notify.
func NewHTTPSink(url string) *HTTPSink {
	rc := retryablehttp.NewClient()
	rc.RetryMax = httpSinkRetryMax
	rc.RetryWaitMin = httpSinkRetryWaitMin
	rc.RetryWaitMax = httpSinkRetryWaitMax
	rc.Logger = nil // не засорять stdout логами ретраев

	return &HTTPSink{
		url:    url,
		client: rc.StandardClient(),
	}
}

// Notify отправляет событие методом POST в формате JSON. При временных
// ошибках сети клиент прозрачно выполнит до httpSinkRetryMax повторов.
// Общее время операции ограничено httpSinkTimeout через context.
func (s *HTTPSink) Notify(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpSinkTimeout)
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
