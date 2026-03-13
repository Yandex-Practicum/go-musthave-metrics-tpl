package agent_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/agent"
)

func TestNewSender(t *testing.T) {
	sender := agent.NewSender("http://localhost:8080")

	if sender == nil {
		t.Fatal("NewSender() returned nil")
	}

}

func TestSendGauge(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Проверяем Content-Type
		if ct := r.Header.Get("Content-Type"); ct != "text/plain" {
			t.Errorf("Expected Content-Type 'text/plain', got '%s'", ct)
		}

		// Проверяем URL path
		expectedPath := "/update/gauge/testGauge/3.14"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := agent.NewSender(server.URL)
	err := sender.SendGauge("testGauge", 3.14)

	if err != nil {
		t.Errorf("SendGauge() failed: %v", err)
	}
}

func TestSendCounter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем URL path
		expectedPath := "/update/counter/testCounter/42"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := agent.NewSender(server.URL)
	err := sender.SendCounter("testCounter", 42)

	if err != nil {
		t.Errorf("SendCounter() failed: %v", err)
	}
}

func TestSendAllMetrics(t *testing.T) {
	receivedMetrics := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Парсим путь: /update/{type}/{name}/{value}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 5 {
			metricType := parts[2]
			metricName := parts[3]
			metricValue := parts[4]
			receivedMetrics[metricName] = metricType + ":" + metricValue
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := agent.NewSender(server.URL)

	// Тестовые данные
	gauges := map[string]float64{
		"gauge1": 1.23,
		"gauge2": 4.56,
	}

	counters := map[string]int64{
		"counter1": 10,
		"counter2": 20,
	}

	err := sender.SendAllMetrics(gauges, counters)
	if err != nil {
		t.Errorf("SendAllMetrics() failed: %v", err)
	}

	// Проверяем что все метрики были отправлены
	expectedMetrics := map[string]string{
		"gauge1":   "gauge:1.23",
		"gauge2":   "gauge:4.56",
		"counter1": "counter:10",
		"counter2": "counter:20",
	}

	for name, expected := range expectedMetrics {
		if received, exists := receivedMetrics[name]; !exists {
			t.Errorf("Metric %s was not sent", name)
		} else if received != expected {
			t.Errorf("Metric %s: expected '%s', got '%s'", name, expected, received)
		}
	}
}

func TestSendMetricServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := agent.NewSender(server.URL)
	err := sender.SendGauge("testGauge", 1.0)

	if err == nil {
		t.Error("Expected error for server error response, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to contain '500', got: %v", err)
	}
}
