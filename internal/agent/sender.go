package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/crypto"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/middleware/hash"
	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/hashicorp/go-retryablehttp"
)

const pollCountName = "PollCount"

type Sender struct {
	baseURL   string
	client    *http.Client
	hashKey   string
	publicKey *rsa.PublicKey
}

func NewSender(baseURL string, hashKey string, publicKey *rsa.PublicKey) *Sender {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 5 * time.Second

	return &Sender{
		baseURL:   baseURL,
		client:    retryClient.StandardClient(),
		hashKey:   hashKey,
		publicKey: publicKey,
	}
}

// encryptIfNeeded шифрует data публичным ключом, если он задан;
// иначе возвращает data без изменений и признак ok=false.
func (s *Sender) encryptIfNeeded(data []byte) ([]byte, bool, error) {
	if s.publicKey == nil {
		return data, false, nil
	}
	encrypted, err := crypto.Encrypt(s.publicKey, data)
	if err != nil {
		return nil, false, fmt.Errorf("шифрование тела: %w", err)
	}
	return encrypted, true, nil
}

func (s *Sender) postjson(m models.Metrics) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// Хеш считаем от исходного JSON — сервер после расшифровки/распаковки
	// сравнивает его с этим значением.
	var hashValue string
	if s.hashKey != "" {
		hashValue = hash.ComputeHMAC(data, s.hashKey)
	}

	// Асимметричное шифрование выполняется до gzip: расшифрованные
	// байты остаются валидным сжатым потоком на сервере.
	payload, encrypted, err := s.encryptIfNeeded(data)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return err
	}
	if _, err := gz.Write(payload); err != nil {
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
	if hashValue != "" {
		req.Header.Set("HashSHA256", hashValue)
	}
	if encrypted {
		req.Header.Set("X-Encrypted", "true")
	}

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

	var hashValue string
	if s.hashKey != "" {
		hashValue = hash.ComputeHMAC(data, s.hashKey)
	}

	payload, encrypted, err := s.encryptIfNeeded(data)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return err
	}
	if _, err := gz.Write(payload); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.baseURL+"/updates/", &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	if hashValue != "" {
		req.Header.Set("HashSHA256", hashValue)
	}
	if encrypted {
		req.Header.Set("X-Encrypted", "true")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка: %s", resp.Status)
	}
	return nil

}
