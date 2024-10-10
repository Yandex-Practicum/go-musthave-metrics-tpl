package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentSendMetrics(t *testing.T) {
	// Создаем фейковый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что URL имеет правильный формат
		assert.Contains(t, r.URL.Path, "/update/", "Incorrect URL path")

		// Проверяем, что запрос содержит правильные метрики
		pathParts := strings.Split(r.URL.Path, "/")
		assert.Len(t, pathParts, 5, "URL path should have 5 parts")

		metricType := pathParts[2]
		metricName := pathParts[3]
		metricValue := pathParts[4]

		// Проверяем тип и значение метрики
		assert.Equal(t, "gauge", metricType, "Metric type should be 'gauge'")
		assert.Equal(t, "testMetric", metricName, "Metric name should be 'testMetric'")
		assert.Equal(t, "42.5", metricValue, "Metric value should be '42.5'")

		// Возвращаем 200 OK
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Переопределим метод sendMetrics, чтобы он отправлял запросы на наш тестовый сервер
	sendMetrics := func(metricType, metricName string, value float64) {
		metricValue := strconv.FormatFloat(value, 'f', -1, 64)
		url := fmt.Sprintf("%s/update/%s/%s/%s", server.URL, metricType, metricName, metricValue)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			t.Fatalf("Error creating request: %v", err)
		}

		req.Header.Set("Content-Type", "text/plain")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Error sending request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK")
	}

	// Вызов отправки метрики
	sendMetrics("gauge", "testMetric", 42.5)
}
