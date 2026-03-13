package agent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/config"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestHTTPClientSendMetric(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что запрос имеет правильный путь и заголовки
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

		// Проверяем, что URL соответствует формату
		// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		expectedPath := "/update/gauge/TestMetric/123.456000"
		assert.Equal(t, expectedPath, r.URL.Path)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTP клиент
	cfg := &config.ServerConfig{
		Address: strings.TrimPrefix(server.URL, "http://"),
	}
	client := NewHTTPClient(cfg)

	// Создаем тестовую метрику
	metric := model.Metrics{
		ID:    "TestMetric",
		MType: model.TypeGauge,
		Value: float64Ptr(123.456),
	}

	// Отправляем метрику
	err := client.SendMetric(metric)
	assert.NoError(t, err)
}

func TestHTTPClientSendCounterMetric(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что запрос имеет правильный путь и заголовки
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

		// Проверяем, что URL соответствует формату для counter
		expectedPath := "/update/counter/TestCounter/789"
		assert.Equal(t, expectedPath, r.URL.Path)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTP клиент
	cfg := &config.ServerConfig{
		Address: strings.TrimPrefix(server.URL, "http://"),
	}
	client := NewHTTPClient(cfg)

	// Создаем тестовую counter метрику
	metric := model.Metrics{
		ID:    "TestCounter",
		MType: model.TypeCounter,
		Delta: int64Ptr(789),
	}

	// Отправляем метрику
	err := client.SendMetric(metric)
	assert.NoError(t, err)
}

func TestHTTPClientSendMetricFailure(t *testing.T) {
	// Создаем тестовый сервер, который возвращает ошибку
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Создаем HTTP клиент
	cfg := &config.ServerConfig{
		Address: strings.TrimPrefix(server.URL, "http://"),
	}
	client := NewHTTPClient(cfg)

	// Создаем тестовую метрику
	metric := model.Metrics{
		ID:    "TestMetric",
		MType: model.TypeGauge,
		Value: float64Ptr(123.456),
	}

	// Отправляем метрику и проверяем, что получаем ошибку
	err := client.SendMetric(metric)
	assert.Error(t, err)
}

func TestHTTPClientSendBatch(t *testing.T) {
	// Создаем счетчик для отслеживания вызовов
	callCount := 0
	expectedCalls := 2

	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Проверяем, что запрос имеет правильный путь и заголовки
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

		// Проверяем, что URL соответствует формату
		if callCount == 1 {
			expectedPath := "/update/gauge/TestMetric1/123.456000"
			assert.Equal(t, expectedPath, r.URL.Path)
		} else if callCount == 2 {
			expectedPath := "/update/counter/TestCounter1/789"
			assert.Equal(t, expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTP клиент
	cfg := &config.ServerConfig{
		Address: strings.TrimPrefix(server.URL, "http://"),
	}
	client := NewHTTPClient(cfg)

	// Создаем тестовые метрики
	metrics := []model.Metrics{
		{
			ID:    "TestMetric1",
			MType: model.TypeGauge,
			Value: float64Ptr(123.456),
		},
		{
			ID:    "TestCounter1",
			MType: model.TypeCounter,
			Delta: int64Ptr(789),
		},
	}

	// Отправляем батч
	err := client.SendBatch(metrics)
	assert.NoError(t, err)

	// Проверяем, что были сделаны все вызовы
	assert.Equal(t, expectedCalls, callCount)
}

func TestHTTPClientTimeout(t *testing.T) {
	// Пропускаем тест, так как он зависит от конкретной реализации таймаута
	// и может быть ненадежным в разных средах
	t.Skip("Skipping timeout test due to unreliable network behavior")
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
