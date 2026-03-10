package agent

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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

func (s *Sender) SendGauge(name string, value float64) error {
	path := fmt.Sprintf("%s/update/gauge/%s/%s", s.baseURL, url.PathEscape(name), strconv.FormatFloat(value, 'f', -1, 64))
	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func (s *Sender) SendCounter(name string, delta int64) error {
	path := fmt.Sprintf("%s/update/counter/%s/%d", s.baseURL, url.PathEscape(name), delta)
	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
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
