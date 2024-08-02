package sender

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
)

// SendMetrics отправляет метрики на сервер
func SendMetrics(metricsData []metrics.Metric) {
	client := resty.New()

	for _, metric := range metricsData {
		url := fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", metric.Type, metric.Name, metric.Value)
		resp, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)

		if err != nil {
			fmt.Printf("Failed to send metric %s: %v\n", metric.Name, err)
			continue
		}

		if resp.StatusCode() != 200 {
			fmt.Printf("Failed to send metric %s: status code %d\n", metric.Name, resp.StatusCode())
		}
	}
}