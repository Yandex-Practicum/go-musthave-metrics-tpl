package provider

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type AgentConfig struct {
	host           string
	pollInterval   time.Duration
	reportInterval time.Duration
	poolCount      int64
}

func NewAgent(host string, pollInterval, reportInterval time.Duration) *AgentConfig {
	return &AgentConfig{
		host:           host,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		poolCount:      0,
	}
}

func (a *AgentConfig) CollectMetrics() map[string]float64 {
	memStats := new(runtime.MemStats)
	runtime.ReadMemStats(memStats)

	metrics := map[string]float64{
		"Alloc":         float64(memStats.Alloc),
		"BuckHashSys":   float64(memStats.BuckHashSys),
		"Frees":         float64(memStats.Frees),
		"GCCPUFraction": memStats.GCCPUFraction,
		"GCSys":         float64(memStats.GCSys),
		"HeapAlloc":     float64(memStats.HeapAlloc),
		"HeapIdle":      float64(memStats.HeapIdle),
		"HeapInuse":     float64(memStats.HeapInuse),
		"HeapObjects":   float64(memStats.HeapObjects),
		"HeapReleased":  float64(memStats.HeapReleased),
		"HeapSys":       float64(memStats.HeapSys),
		"LastGC":        float64(memStats.LastGC),
		"Lookups":       float64(memStats.Lookups),
		"MCacheInuse":   float64(memStats.MCacheInuse),
		"MCacheSys":     float64(memStats.MCacheSys),
		"MSpanInuse":    float64(memStats.MSpanInuse),
		"MSpanSys":      float64(memStats.MSpanSys),
		"Mallocs":       float64(memStats.Mallocs),
		"NextGC":        float64(memStats.NextGC),
		"NumForcedGC":   float64(memStats.NumForcedGC),
		"NumGC":         float64(memStats.NumGC),
		"OtherSys":      float64(memStats.OtherSys),
		"PauseTotalNs":  float64(memStats.PauseTotalNs),
		"StackInuse":    float64(memStats.StackInuse),
		"StackSys":      float64(memStats.StackSys),
		"Sys":           float64(memStats.Sys),
		"TotalAlloc":    float64(memStats.TotalAlloc),
		"RandomValue":   rand.Float64() * 100,
	}

	return metrics
}

func (a *AgentConfig) SendMetrics(metricType, metricName string, value float64) {
	metricValue := strconv.FormatFloat(value, 'f', -1, 64)
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.host, metricType, metricName, metricValue)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	fmt.Printf("Metrics %s (%s) with value %s sent successfully\n", metricName, metricType, metricValue)
}

func (a *AgentConfig) Start(ctx context.Context) {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Agent shutting down gracefully...")
			return
		case <-pollTicker.C:
			a.poolCount++
			metrics := a.CollectMetrics()
			metrics["PoolCount"] = float64(a.poolCount)
			fmt.Println("Metrics collected:", metrics)
		case <-reportTicker.C:
			metrics := a.CollectMetrics()
			metrics["PoolCount"] = float64(a.poolCount)
			fmt.Println("Metrics collected for report:", metrics)

			for name, value := range metrics {
				a.SendMetrics("gauge", name, value)
			}
			a.SendMetrics("counter", "PoolCount", float64(a.poolCount))
		}
	}
}
