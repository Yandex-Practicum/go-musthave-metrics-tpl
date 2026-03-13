package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Sender struct {
	client  *http.Client
	baseURL string
}

func NewSender(baseURL string) *Sender {
	return &Sender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// SendGauge отправляет gauge метрику
func (s *Sender) SendGauge(name string, value float64) error {
	url := fmt.Sprintf("%s/update/gauge/%s/%s",
		s.baseURL, name, strconv.FormatFloat(value, 'f', -1, 64))
	return s.sendMetric(url, "gauge", name)
}

// SendCounter отправляет counter метрику
func (s *Sender) SendCounter(name string, value int64) error {
	url := fmt.Sprintf("%s/update/counter/%s/%d", s.baseURL, name, value)
	return s.sendMetric(url, "counter", name)
}

// SendAllMetrics отправляет все метрики на сервер
func (s *Sender) SendAllMetrics(gauge map[string]float64, counter map[string]int64) error {
	totalMetrics := len(gauge) + len(counter)
	sentMetrics := 0

	log.Printf("Sending %d gauge metrics and %d counter metrics", len(gauge), len(counter))

	// Отправляем все gauge метрики
	for name, value := range gauge {
		if err := s.SendGauge(name, value); err != nil {
			return fmt.Errorf("failed to send gauge %s: %w", name, err)
		}
		sentMetrics++
	}

	// Отправляем все counter метрики
	for name, value := range counter {
		if err := s.SendCounter(name, value); err != nil {
			return fmt.Errorf("failed to send counter %s: %w", name, err)
		}
		sentMetrics++
	}

	log.Printf("Successfully sent %d/%d metrics", sentMetrics, totalMetrics)
	return nil
}

func (s *Sender) sendMetric(url, metricType, metricName string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	//Устанавливаем требуемый заголовок
	req.Header.Set("Content-Type", "text/plain")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d for %s %s",
			resp.StatusCode, metricType, metricName)
	}

	log.Printf("Sent %s metric: %s", metricType, metricName)
	return nil
}

func addHashHeader(req *http.Request, body []byte, key string) {
	if key == "" {
		return
	}
	hash := ComputeHMAC(body, key)
	req.Header.Set("HashSHA256", hash)
}
