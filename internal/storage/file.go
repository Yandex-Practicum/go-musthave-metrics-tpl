package storage

import (
	"context"
	"encoding/json"
	"os"
	"time"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/rs/zerolog/log"
)

func SaveToFile(s *MemoryStorage, filename string) error {
	var metrics []models.Metrics

	for name, value := range s.GetAllGauges(context.Background()) {
		v := value
		metrics = append(metrics, models.Metrics{ID: name, MType: "gauge", Value: &v})
	}

	for name, value := range s.GetAllCounters(context.Background()) {
		d := value
		metrics = append(metrics, models.Metrics{ID: name, MType: "counter", Delta: &d})
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(metrics)
}

func LoadFromFile(s *MemoryStorage, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var metrics []models.Metrics
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return err
	}

	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			if m.Value != nil {
				s.UpdateGauge(context.Background(), m.ID, *m.Value)
			}
		case "counter":
			if m.Delta != nil {
				s.UpdateCounter(context.Background(), m.ID, *m.Delta)
			}
		}
	}

	return nil

}

func RunSaver(s *MemoryStorage, filename string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		if err := SaveToFile(s, filename); err != nil {
			log.Error().Err(err).Msg("Ошибка сохранения метрик")
		}
	}
}
