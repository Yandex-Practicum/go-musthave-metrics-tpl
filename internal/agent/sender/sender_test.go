package sender

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
)

// toFloat64Pointer преобразует значение float64 в указатель на float64
func toFloat64Pointer(value float64) *float64 {
	return &value
}

// toInt64Pointer преобразует значение int64 в указатель на int64
func toInt64Pointer(value int64) *int64 {
	return &value
}

func TestSendMetrics(t *testing.T) {
	tests := []struct {
		name        string
		metricsData []metrics.Metrics
		statusCode  int
		expectError bool
	}{
		{
			name: "Valid metrics",
			metricsData: []metrics.Metrics{
				{MType: "gauge", ID: "metric1", Value: toFloat64Pointer(1.23)},
				{MType: "counter", ID: "metric2", Delta: toInt64Pointer(10)},
				{MType: "counter", ID: "metric3", Delta: toInt64Pointer(0)},
			},
			statusCode:  http.StatusOK,
			expectError: false,
		},
		// {
		//     name: "Server error",
		//     metricsData: []metrics.Metrics{
		//         {MType: "gaug", ID: "metric1", Value: nil},
		//     },
		//     statusCode:  http.StatusInternalServerError,
		//     expectError: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок-сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем содержимое запроса
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "text/plain" {
					t.Errorf("Expected Content-Type text/plain, got %s", r.Header.Get("Content-Type"))
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := resty.New()
			for _, metric := range tt.metricsData {
				url := fmt.Sprintf("%s/update/%s/%s/%v", server.URL, metric.MType, metric.ID, metric.Value)
				resp, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(url)

				if tt.expectError {
					if err == nil {
						t.Errorf("Expected error but got none")
					}
				} else {
					if err != nil {
						t.Errorf("Failed to send metric %s: %v", metric.ID, err)
					}
					if resp.StatusCode() != tt.statusCode {
						t.Errorf("Expected status code %d but got %d", tt.statusCode, resp.StatusCode())
					}
				}
			}
		})
	}
}
