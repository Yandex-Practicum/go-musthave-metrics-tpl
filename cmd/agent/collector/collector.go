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
	poolCount      int64
	collector      *metrics.Collector
	httpClient     *httpclient.HTTPClient
}

func NewAgent(host string, pollInterval, reportInterval time.Duration) *AgentConfig {
	return &AgentConfig{
		host:           host,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		poolCount:      0,
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
			a.poolCount++
			collectedMetrics := a.collector.CollectMetrics()
			collectedMetrics["PoolCount"] = float64(a.poolCount)
			fmt.Println("Metrics collected:", collectedMetrics)
		case <-reportTicker.C:
			collectedMetrics := a.collector.CollectMetrics()
			collectedMetrics["PoolCount"] = float64(a.poolCount)
			fmt.Println("Metrics collected for report:", collectedMetrics)

			for name, value := range collectedMetrics {
				a.httpClient.SendMetrics("gauge", name, value)
			}
			a.httpClient.SendMetrics("counter", "PoolCount", float64(a.poolCount))
		}
	}
}
