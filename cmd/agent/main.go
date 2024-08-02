package main

import (
    "fmt"
    "sync"
    "time"

	"github.com/vova4o/yandexadv/internal/agent/collector"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
	"github.com/vova4o/yandexadv/internal/agent/sender"
)

// change this later to a config stuct
var (
    pollInterval   = 2 * time.Second
    reportInterval = 10 * time.Second
    pollCount      int64
    metricsData    []metrics.Metric
    metricsMutex   sync.Mutex
)

func main() {
    tickerPoll := time.NewTicker(pollInterval)
    tickerReport := time.NewTicker(reportInterval)

    for {
        select {
        case <-tickerPoll.C:
            pollCount++
            metricsMutex.Lock()
            metricsData = collector.CollectMetrics(pollCount)
            metricsMutex.Unlock()

        case <-tickerReport.C:
            metricsMutex.Lock()
            sender.SendMetrics(metricsData)
            metricsMutex.Unlock()
            fmt.Println("Sent metrics")
        }
    }
}