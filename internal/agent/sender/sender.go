package sender

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
)

// SendMetrics отправляет метрики на сервер
func SendMetrics(address string, metricsData []metrics.Metrics) {
	client := resty.New()

	for _, metric := range metricsData {
		var url string
		if metric.Value == nil {
			url = fmt.Sprintf("http://%s/update/%s/%s/%v", address, metric.MType, metric.ID, *metric.Delta)
		} else {
			url = fmt.Sprintf("http://%s/update/%s/%s/%v", address, metric.MType, metric.ID, *metric.Value)
		}

		resp, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)

		if err != nil {
			log.Printf("Failed to send metric %s: %v\n", metric.ID, err)
			continue
		}

		if resp.StatusCode() != 200 {
			log.Printf("Failed to send metric %s: status code %d\n", metric.ID, resp.StatusCode())
		}
	}
}

// SendMetricsJSON отправляет метрики на сервер в формате JSON
func SendMetricsJSON(address string, metricsData []metrics.Metrics) {
	client := resty.New()

	for _, metric := range metricsData {
		url := fmt.Sprintf("http://%s/update/", address)

		// Сериализация метрики в JSON
		jsonData, err := json.Marshal(metric)
		if err != nil {
			log.Printf("Failed to marshal metric %s: %v\n", metric.ID, err)
			continue
		}

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(jsonData).
			Post(url)

		if err != nil {
			log.Printf("Failed to send metric %s: %v\n", metric.ID, err)
			continue
		}

		if resp.StatusCode() != 200 {
			log.Printf("Failed to send metric %s: status code %d\n", metric.ID, resp.StatusCode())
		}
	}
}
