package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/config"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
)

type Agent struct {
	cfg         *config.Config
	metricsChan chan model.Metrics
	client      *HTTPClient
	wg          sync.WaitGroup
}

func NewAgent(cfg *config.Config) *Agent {
	return &Agent{
		cfg:         cfg,
		metricsChan: make(chan model.Metrics, 1000),
		client:      NewHTTPClient(cfg),
	}
}

func (a *Agent) runWorkers(ctx context.Context) {
	defer a.wg.Done()

	sem := make(chan struct{}, a.cfg.RateLimit)
	defer close(sem)

	batchSize := 20                        // или задаётся из конфига
	batchTimeout := 500 * time.Millisecond // максимальное время ожидания для наполнения батча

	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-a.metricsChan:
			if !ok {
				return
			}

			// Собираем батч
			batch := make([]model.Metrics, 0, batchSize)
			batch = append(batch, m)

		collectLoop:
			for len(batch) < batchSize {
				select {
				case m2, ok := <-a.metricsChan:
					if !ok {
						break collectLoop
					}
					batch = append(batch, m2)
				case <-time.After(batchTimeout):
					break collectLoop
				case <-ctx.Done():
					return
				}
			}

			// Запускаем горутину отправки батча (ограничиваем sem)
			sem <- struct{}{}
			go func(metricsBatch []model.Metrics) {
				defer func() { <-sem }()
				err := a.client.SendBatch(metricsBatch)
				if err != nil {
					log.Printf("Failed to send batch: %v", err)
				}
			}(batch)
		}
	}
}

func (a *Agent) sendGauge(name string, value float64) {
	a.metricsChan <- model.Metrics{
		ID:    name,
		MType: "gauge",
		Value: &value,
	}
}

func (a *Agent) sendCounter(name string, value int64) {
	a.metricsChan <- model.Metrics{
		ID:    name,
		MType: "counter",
		Delta: &value,
	}
}

type Client interface {
	SendMetric(metric model.Metrics) error
	SendBatch(metrics []model.Metrics) error
}

func SendGzipJSON(url string, jsonData []byte) error {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(jsonData); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip") // тело запроса в gzip
	req.Header.Set("Accept-Encoding", "gzip")  // ожидаем gzipped ответ

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body

	// Распаковываем gzip ответ, если он есть
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Читаем тело ответа (можно декодировать JSON дальше)
	_, err = io.ReadAll(reader)
	return err
}
