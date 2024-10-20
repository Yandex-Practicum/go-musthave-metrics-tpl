package collector

import (
	"context"
	"fmt"
	"time"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/agent/httpclient"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/agent/metrics"
)

type AgentConfig struct {
	host           string
	pollInterval   time.Duration
	reportInterval time.Duration
	PoolCount      int64
	collector      *metrics.Collector
	httpClient     *httpclient.HTTPClient
}

func NewAgent(host string, pollInterval, reportInterval time.Duration) *AgentConfig {
	return &AgentConfig{
		host:           host,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		PoolCount:      0,
		collector:      metrics.NewMetricsCollector(),
		httpClient:     httpclient.NewHTTPClient(host),
	}
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
			a.PoolCount++
			collectedMetrics := a.collector.CollectMetrics()
			collectedMetrics = append(collectedMetrics, metrics.GenerateJSON(metrics.Metrics{ID: "PollCount", MType: "counter", Delta: &a.PoolCount}))
			var jsonSlice []string
			for _, m := range collectedMetrics {
				jsonSlice = append(jsonSlice, string(m))
			}
			fmt.Println("Metrics collected:", jsonSlice)
		case <-reportTicker.C:
			collectedMetrics := a.collector.CollectMetrics()
			collectedMetrics = append(collectedMetrics, metrics.GenerateJSON(metrics.Metrics{ID: "PollCount", MType: "counter", Delta: &a.PoolCount}))
			for _, data := range collectedMetrics {
				a.httpClient.SendMetrics(data)
				fmt.Println("Reported: ", string(data))
			}
		}
	}
}
