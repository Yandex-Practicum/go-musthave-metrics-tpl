package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/vova4o/yandexadv/internal/agent/collector"
	"github.com/vova4o/yandexadv/internal/agent/flags"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
	"github.com/vova4o/yandexadv/internal/agent/sender"
	"github.com/vova4o/yandexadv/package/logger"
)

// change this later to a config stuct
var (
	pollCount    int64
	metricsData  []metrics.Metrics
	metricsMutex sync.Mutex
)

func main() {
	config := flags.NewConfig()

	logger, err := logger.NewLogger("info", config.AgenLogFileName)
	if err != nil {
		fmt.Println("Error creating logger")
		return
	}

	logger.Info("Starting agent")
	logger.Info("Server address: " + config.ServerAddress)

	tickerPoll := time.NewTicker(config.PollInterval)
	tickerReport := time.NewTicker(config.ReportInterval)

	for {
		select {
		case <-tickerPoll.C:
			pollCount++
			metricsMutex.Lock()
			metricsData = collector.CollectMetrics(pollCount)
			metricsMutex.Unlock()

		case <-tickerReport.C:
			metricsMutex.Lock()
			// sender.SendMetrics(config.ServerAddress, metricsData)
			// sender.SendMetricsJSON(config.ServerAddress, metricsData)
			sender.SendMetricsBatch(config, metricsData)
			metricsMutex.Unlock()
		}
	}
}
