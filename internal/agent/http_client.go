package agent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/config"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
)

type HTTPClient struct {
	cfg    *config.ServerConfig
	client *http.Client
}

func (c *HTTPClient) SendBatch(metricsBatch []model.Metrics) error {
	// Для простоты реализации отправим по одной метрике
	for _, metric := range metricsBatch {
		if err := c.SendMetric(metric); err != nil {
			return err
		}
	}
	return nil
}

func NewHTTPClient(cfg *config.ServerConfig) *HTTPClient {
	return &HTTPClient{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *HTTPClient) SendMetric(m model.Metrics) error {
	// Формат данных — http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	var url string
	if m.MType == model.TypeGauge {
		url = fmt.Sprintf("http://%s/update/%s/%s/%.6f", c.cfg.Address, m.MType, m.ID, *m.Value)
	} else if m.MType == model.TypeCounter {
		url = fmt.Sprintf("http://%s/update/%s/%s/%d", c.cfg.Address, m.MType, m.ID, *m.Delta)
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	// Устанавливаем заголовок
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}
