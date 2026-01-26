package agent

import (
	"log"
	"math/rand"
	"runtime"
	"sync"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
)

type Collector struct {
	mu      sync.RWMutex
	gauge   map[string]float64
	counter map[string]int64
}

func NewCollector() *Collector {
	return &Collector{
		mu:      sync.RWMutex{},
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// UpdateMetrics собирает runtime-метрики и обновляет RandomValue и PollCount
func (c *Collector) UpdateMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Gauge метрики из runtime
	c.gauge["Alloc"] = float64(m.Alloc)
	c.gauge["BuckHashSys"] = float64(m.BuckHashSys)
	c.gauge["Frees"] = float64(m.Frees)
	c.gauge["GCCPUFraction"] = m.GCCPUFraction
	c.gauge["GCSys"] = float64(m.GCSys)
	c.gauge["HeapAlloc"] = float64(m.HeapAlloc)
	c.gauge["HeapIdle"] = float64(m.HeapIdle)
	c.gauge["HeapInuse"] = float64(m.HeapInuse)
	c.gauge["HeapObjects"] = float64(m.HeapObjects)
	c.gauge["HeapReleased"] = float64(m.HeapReleased)
	c.gauge["HeapSys"] = float64(m.HeapSys)
	c.gauge["LastGC"] = float64(m.LastGC)
	c.gauge["Lookups"] = float64(m.Lookups)
	c.gauge["MCacheInuse"] = float64(m.MCacheInuse)
	c.gauge["MCacheSys"] = float64(m.MCacheSys)
	c.gauge["MSpanInuse"] = float64(m.MSpanInuse)
	c.gauge["MSpanSys"] = float64(m.MSpanSys)
	c.gauge["Mallocs"] = float64(m.Mallocs)
	c.gauge["NextGC"] = float64(m.NextGC)
	c.gauge["NumForcedGC"] = float64(m.NumForcedGC)
	c.gauge["NumGC"] = float64(m.NumGC)
	c.gauge["OtherSys"] = float64(m.OtherSys)
	c.gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	c.gauge["StackInuse"] = float64(m.StackInuse)
	c.gauge["StackSys"] = float64(m.StackSys)
	c.gauge["Sys"] = float64(m.Sys)
	c.gauge["TotalAlloc"] = float64(m.TotalAlloc)

	// RandomValue (тип gauge)
	c.gauge["RandomValue"] = rand.Float64()

	// Counter метрики
	c.counter["PollCount"] = c.counter["PollCount"] + 1
}

// GetGauges возвращает копию всех gauge метрик
func (c *Collector) GetGauges() map[string]float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]float64, len(c.gauge))
	for k, v := range c.gauge {
		result[k] = v
	}
	return result
}

// GetCounters возвращает копию всех counter метрик
func (c *Collector) GetCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.counter))
	for k, v := range c.counter {
		result[k] = v
	}
	return result
}

// GetMetricsCount возвращает количество метрик
func (c *Collector) GetMetricsCount() (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.gauge), len(c.counter)
}

// SendMetrics отправляет метрики через HTTP клиент
func (c *Collector) SendMetrics(client *HTTPClient, addr string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Отправляем gauge метрики
	for name, value := range c.gauge {
		metric := model.Metrics{
			ID:    name,
			MType: model.TypeGauge,
			Value: &value,
		}
		if err := client.SendMetric(metric); err != nil {
			log.Printf("Failed to send gauge metric %s: %v", name, err)
		}
	}

	// Отправляем counter метрики
	for name, value := range c.counter {
		metric := model.Metrics{
			ID:    name,
			MType: model.TypeCounter,
			Delta: &value,
		}
		if err := client.SendMetric(metric); err != nil {
			log.Printf("Failed to send counter metric %s: %v", name, err)
		}
	}

	return nil
}
