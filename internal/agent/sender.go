package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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
		client:  &http.Client{},
	}
}

func (s *Sender) postjson(m models.Metrics) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	resp, err := s.client.Post(s.baseURL+"/update/", "application/json", bytes.NewBuffer(data))
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
