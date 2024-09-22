package sender

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/vova4o/yandexadv/internal/agent/flags"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
)

const maxRetries = 3
const retryDelay = 1 * time.Second

// CompressData сжимает данные с использованием gzip
func CompressData(data []byte) ([]byte, error) {
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

// ServerSupportsGzip проверяет, поддерживает ли сервер gzip-сжатие
func ServerSupportsGzip(address string) bool {
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

// calculateHash вычисляет HMAC-SHA256 хэш из данных и ключа
func calculateHash(data, key []byte) string {
    h := hmac.New(sha256.New, key)
    h.Write(data)
    return hex.EncodeToString(h.Sum(nil))
}

// SendMetricsBatch отправляет метрики на сервер пакетом
func SendMetricsBatch(cfg *flags.Config, metricsData []metrics.Metrics) {
	address := cfg.ServerAddress
	key := cfg.SecretKey

	client := resty.New()
	useGzip := ServerSupportsGzip(address)

	url := fmt.Sprintf("http://%s/updates", address)

	// Сериализация метрик в JSON
	jsonData, err := json.Marshal(metricsData)
	if err != nil {
		log.Printf("Failed to marshal metrics: %v\n", err)
		return
	}

	var hash string
	if key != "" {
		hash = calculateHash(jsonData, []byte(key))
	}


	request := client.R().SetHeader("Content-Type", "application/json")
	request.SetHeader("HashSHA256", hash)

	if useGzip {
		request.SetHeader("Content-Encoding", "gzip")
		compressedData, err := CompressData(jsonData)
		if err != nil {
			log.Printf("Failed to compress data for metrics: %v\n", err)
			return
		}
		request.SetBody(compressedData)
	} else {
		request.SetBody(jsonData)
	}

	if err := sendWithRetry(request, url); err != nil {
		log.Printf("Failed to send metrics: %v\n", err)
	}
}

// SendMetrics отправляет метрики на сервер
func SendMetrics(address string, metricsData []metrics.Metrics) {
	client := resty.New()
	useGzip := ServerSupportsGzip(address)

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
			compressedData, err := CompressData([]byte(url))
			if err != nil {
				log.Printf("Failed to compress data for metric %s: %v\n", metric.ID, err)
				continue
			}
			request.SetBody(compressedData)
		} else {
			request.SetBody(url)
		}

		if err := sendWithRetry(request, url); err != nil {
			log.Printf("Failed to send metric %s: %v\n", metric.ID, err)
		}
	}
}

// SendMetricsJSON отправляет метрики на сервер в формате JSON
func SendMetricsJSON(address string, metricsData []metrics.Metrics) {
	client := resty.New()
	useGzip := ServerSupportsGzip(address)

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
			compressedData, err := CompressData(jsonData)
			if err != nil {
				log.Printf("Failed to compress data for metric %s: %v\n", metric.ID, err)
				continue
			}
			request.SetBody(compressedData)
		} else {
			request.SetBody(jsonData)
		}

		if err := sendWithRetry(request, url); err != nil {
			log.Printf("Failed to send metric %s: %v\n", metric.ID, err)
		}
	}
}

// sendWithRetry отправляет запрос с повторными попытками в случае ошибки
func sendWithRetry(request *resty.Request, url string) error {
    delay := retryDelay
	for i := 0; i < maxRetries; i++ {
        resp, err := request.Post(url)
        if err != nil {
            log.Printf("Failed to send request: %v\n", err)
        } else if resp.StatusCode() == 200 {
            return nil
        } else {
            log.Printf("Failed to send request: status code %d\n", resp.StatusCode())
        }

        time.Sleep(delay)
        delay += 2 * time.Second
    }
    return fmt.Errorf("failed to send request after %d attempts", maxRetries)
}