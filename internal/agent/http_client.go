package agent

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/config"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
)

type HTTPClient struct {
	cfg    *config.Config
	client *http.Client
}

func (c *HTTPClient) SendBatch(metricsBatch []model.Metrics) any {
	panic("unimplemented")
}

func NewHTTPClient(cfg *config.Config) *HTTPClient {
	return &HTTPClient{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *HTTPClient) SendMetric(m model.Metrics) error {
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.cfg.Address, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if c.cfg.Key != "" {
		hash := sha256.Sum256(append(body, []byte(c.cfg.Key)...))
		req.Header.Set("HashSHA256", fmt.Sprintf("%x", hash))
	}

	c.client.Do(req)
	// ... обработка ответа
	return nil
}
