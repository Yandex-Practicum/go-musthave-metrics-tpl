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

var (
	pollCount    int64
	metricsMutex sync.Mutex
)

// AllMetrics структура для хранения всех метрик
type AllMetrics struct {
	RuntimeMetrics    []metrics.Metrics `json:"runtime_metrics"`
	AdditionalMetrics []metrics.Metrics `json:"additional_metrics"`
}

func main() {
	config := flags.NewConfig()

	logger, err := logger.NewLogger("info", config.AgenLogFileName)
	if err != nil {
		fmt.Println("Error creating logger")
		return
	}

	logger.Info("Starting agent")
	logger.Info("Server address: " + config.ServerAddress)
	logger.Info("Secret key: " + config.SecretKey)
	logger.Info("Rate limit: " + fmt.Sprintf("%d", config.RateLimit))

	tickerPoll := time.NewTicker(config.PollInterval)
	tickerReport := time.NewTicker(config.ReportInterval)

	if config.RateLimit == 0 {
		// Старый способ отправки метрик
		go func() {
			for {
				select {
				case <-tickerPoll.C:
					pollCount++
					metricsMutex.Lock()
					runtimeMetrics := collector.CollectMetrics(pollCount)
					additionalMetrics := collector.CollectCPUAndMemlMetrics(pollCount)
					metricsMutex.Unlock()

					allMetrics := append(runtimeMetrics, additionalMetrics...)
					sender.SendMetricsBatch(config, allMetrics)

				case <-tickerReport.C:
					metricsMutex.Lock()
					runtimeMetrics := collector.CollectMetrics(pollCount)
					additionalMetrics := collector.CollectCPUAndMemlMetrics(pollCount)
					metricsMutex.Unlock()

					allMetrics := append(runtimeMetrics, additionalMetrics...)
					sender.SendMetricsBatch(config, allMetrics)
				}
			}
		}()

		select {}
	} else {
		// Новый способ отправки метрик с использованием горутин и каналов
		metricsChan := make(chan AllMetrics, config.RateLimit)
		var wg sync.WaitGroup

		// Запускаем воркеры
		for i := 0; i < config.RateLimit; i++ {
			wg.Add(1)
			go worker(metricsChan, &wg, config)
		}

		// Горутина для сбора runtime метрик
		go func() {
			for {
				select {
				case <-tickerPoll.C:
					pollCount++
					metricsMutex.Lock()
					runtimeMetrics := collector.CollectMetrics(pollCount)
					metricsMutex.Unlock()

					metricsChan <- AllMetrics{RuntimeMetrics: runtimeMetrics}
				}
			}
		}()

		// Горутина для сбора дополнительных метрик
		go func() {
			for {
				select {
				case <-tickerPoll.C:
					metricsMutex.Lock()
					additionalMetrics := collector.CollectCPUAndMemlMetrics(pollCount)
					metricsMutex.Unlock()

					metricsChan <- AllMetrics{AdditionalMetrics: additionalMetrics}
				}
			}
		}()

		// Горутина для отправки метрик на сервер
		go func() {
			for {
				select {
				case <-tickerReport.C:
					metricsMutex.Lock()
					var combinedMetrics AllMetrics
					for i := 0; i < config.RateLimit; i++ {
						metrics := <-metricsChan
						combinedMetrics.RuntimeMetrics = append(combinedMetrics.RuntimeMetrics, metrics.RuntimeMetrics...)
						combinedMetrics.AdditionalMetrics = append(combinedMetrics.AdditionalMetrics, metrics.AdditionalMetrics...)
					}
					metricsMutex.Unlock()

					allMetrics := append(combinedMetrics.RuntimeMetrics, combinedMetrics.AdditionalMetrics...)
					sender.SendMetricsBatch(config, allMetrics)
				}
			}
		}()

		select {}
	}
}

func worker(metricsChan chan AllMetrics, wg *sync.WaitGroup, config *flags.Config) {
	defer wg.Done()
	for metrics := range metricsChan {
		allMetrics := append(metrics.RuntimeMetrics, metrics.AdditionalMetrics...)
		sender.SendMetricsBatch(config, allMetrics)
	}
}
