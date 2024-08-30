package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
)

// compressData сжимает данные с использованием gzip
func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// serverSupportsGzip проверяет, поддерживает ли сервер gzip-сжатие
func serverSupportsGzip(address string) bool {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept-Encoding", "gzip").
		Get(fmt.Sprintf("http://%s", address))

	if err != nil {
		log.Printf("Failed to check gzip support: %v\n", err)
		return false
	}

	return resp.Header().Get("Content-Encoding") == "gzip"
}

// SendMetrics отправляет метрики на сервер
func SendMetrics(address string, metricsData []metrics.Metrics) {
	client := resty.New()
	useGzip := serverSupportsGzip(address)

	for _, metric := range metricsData {
		var url string
		if metric.Value == nil {
			url = fmt.Sprintf("http://%s/update/%s/%s/%v", address, metric.MType, metric.ID, *metric.Delta)
		} else {
			url = fmt.Sprintf("http://%s/update/%s/%s/%v", address, metric.MType, metric.ID, *metric.Value)
		}

		request := client.R().SetHeader("Content-Type", "text/plain")

		if useGzip {
			request.SetHeader("Content-Encoding", "gzip")
			compressedData, err := compressData([]byte(url))
			if err != nil {
				log.Printf("Failed to compress data for metric %s: %v\n", metric.ID, err)
				continue
			}
			request.SetBody(compressedData)
		} else {
			request.SetBody(url)
		}

		resp, err := request.Post(url)

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
	useGzip := serverSupportsGzip(address)

	for _, metric := range metricsData {
		url := fmt.Sprintf("http://%s/update/", address)

		// Сериализация метрики в JSON
		jsonData, err := json.Marshal(metric)
		if err != nil {
			log.Printf("Failed to marshal metric %s: %v\n", metric.ID, err)
			continue
		}

		request := client.R().SetHeader("Content-Type", "application/json")

		if useGzip {
			request.SetHeader("Content-Encoding", "gzip")
			compressedData, err := compressData(jsonData)
			if err != nil {
				log.Printf("Failed to compress data for metric %s: %v\n", metric.ID, err)
				continue
			}
			request.SetBody(compressedData)
		} else {
			request.SetBody(jsonData)
		}

		resp, err := request.Post(url)

		if err != nil {
			log.Printf("Failed to send metric %s: %v\n", metric.ID, err)
			continue
		}

		if resp.StatusCode() != 200 {
			log.Printf("Failed to send metric %s: status code %d\n", metric.ID, resp.StatusCode())
		}
	}
}
