package agent

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
)

type Collector struct {
	mu        *sync.Mutex
	gauge     map[string]float64
	counter   map[string]int64
	buffer    []model.Metrics
	BatchSize int
	client    *http.Client // Добавлено поле для HTTP-клиента
	endpoint  string       // Добавлено поле для адреса сервера
}

func (c *Collector) UpdateMetrics() {
	panic("unimplemented")
}

type AgentConfig struct {
	BatchSize     int           `env:"BATCH_SIZE" default:"50"`
	FlushInterval time.Duration `env:"FLUSH_INTERVAL" default:"5s"`
}

func NewCollector(batchSize int, client *http.Client, endpoint string) *Collector {
	return &Collector{
		BatchSize: batchSize,
		mu:        &sync.Mutex{},
		gauge:     make(map[string]float64),
		counter:   make(map[string]int64),
		client:    client,   // Инициализация клиента
		endpoint:  endpoint, // Инициализация адреса сервера
	}
}

func (c *Collector) UdateMetrics() {
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

	//RandomValue (тип gauge) — обновляемое произвольное значение
	c.gauge["RandomValue"] = rand.Float64()

	// Counter метрики
	c.counter["PollCount"]++
}

// GetGauges возвращает копию всех gauge метрик
func (c *Collector) GetGauges() map[string]float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make(map[string]float64)
	for k, v := range c.counter {
		result[k] = float64(v)
	}
	return result
}

// GetCounters возвращает копию всех counter метрик
func (c *Collector) GetCounters() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make(map[string]int64)
	for k, v := range c.counter {
		result[k] = v
	}
	return result
}

// GetMetricsCount возвращает количество метрик
func (c *Collector) GetMetricsCount() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.gauge), len(c.counter)
}

func (c *Collector) AddMetric(m model.Metrics) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer = append(c.buffer, m)
	if len(c.buffer) >= c.BatchSize {
		go c.sendBatch()
	}
}

func (c *Collector) sendBatch() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.buffer) == 0 {
		return
	}

	body, err := json.Marshal(c.buffer)
	if err != nil {
		// Добавить обработку ошибки
		return
	}

	compressed := compress(body)

	req, err := http.NewRequest("POST", c.endpoint+"/updates", bytes.NewReader(compressed))
	if err != nil {
		// Добавить обработку ошибки
		return
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		// Добавить обработку ошибки
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.buffer = c.buffer[:0]
	}
}

func (c *Collector) StartFlusher(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			c.mu.Lock()
			if len(c.buffer) > 0 {
				go c.sendBatch()
			}
			c.mu.Unlock()
		}
	}()
}
