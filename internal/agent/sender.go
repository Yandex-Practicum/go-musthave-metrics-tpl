package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
)

const pollCountName = "PollCount"

type Sender struct {
	baseURL string
	client  *http.Client
}

func NewSender(baseURL string) *Sender {
	return &Sender{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (s *Sender) postjson(m models.Metrics) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return err
	}

	if _, err := gz.Write(data); err != nil {
		return err
	}

	if err := gz.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.baseURL+"/update/", &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func (s *Sender) SendGauge(name string, value float64) error {
	return s.postjson(models.Metrics{
		ID:    name,
		MType: models.Gauge,
		Value: &value,
	})
}

func (s *Sender) SendCounter(name string, delta int64) error {
	return s.postjson(models.Metrics{
		ID:    name,
		MType: models.Counter,
		Delta: &delta,
	})
}

func (s *Sender) SendAll(gauges []GaugeMetric, pollCountDelta int64) error {
	for _, g := range gauges {
		if err := s.SendGauge(g.Name, g.Value); err != nil {
			return fmt.Errorf("send gauge %s: %w", g.Name, err)
		}
	}
	if err := s.SendCounter(pollCountName, pollCountDelta); err != nil {
		return fmt.Errorf("send counter %s: %w", pollCountName, err)
	}
	return nil
}
