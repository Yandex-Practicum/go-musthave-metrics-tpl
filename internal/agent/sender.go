package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/retry"
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

	compressed := buf.Bytes()

	return retry.Do(context.Background(), func() error {
		req, err := http.NewRequest(http.MethodPost, s.baseURL+"/update/", bytes.NewReader(compressed))
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
	})
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

func (s *Sender) SendBatch(gauge []GaugeMetric, pollCountDelta int64) error {
	if len(gauge) == 0 && pollCountDelta == 0 {
		return nil
	}

	metrics := make([]models.Metrics, 0, len(gauge)+1)

	for _, g := range gauge {
		v := g.Value
		metrics = append(metrics, models.Metrics{
			ID:    g.Name,
			MType: models.Gauge,
			Value: &v,
		})
	}
	delta := pollCountDelta
	metrics = append(metrics, models.Metrics{
		ID:    pollCountName,
		MType: models.Counter,
		Delta: &delta,
	})

	data, err := json.Marshal(metrics)
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

	compressed := buf.Bytes()

	return retry.Do(context.Background(), func() error {
		req, err := http.NewRequest(http.MethodPost, s.baseURL+"/updates/", bytes.NewReader(compressed))
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
			return fmt.Errorf("ошибка: %s", resp.Status)
		}
		return nil
	})
}
