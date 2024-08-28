package sender

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
)

// SendMetrics отправляет метрики на сервер
func SendMetrics(address string, metricsData []metrics.Metric) {
	client := resty.New()

	for _, metric := range metricsData {
		url := fmt.Sprintf("http://%s/update/%s/%s/%v", address, metric.Type, metric.Name, metric.Value)
		resp, err := client.R().
			SetHeader("Content-Type", "text/plain").
			Post(url)

		if err != nil {
			log.Printf("Failed to send metric %s: %v\n", metric.Name, err)
			continue
		}

		if resp.StatusCode() != 200 {
			log.Printf("Failed to send metric %s: status code %d\n", metric.Name, resp.StatusCode())
		}
	}
}
