package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/agent"
)

// Конфигурация агента
type AgentConfig struct {
	ServerAddress  string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

// Константы по умолчанию согласно требованиям
const (
	defaultPollInterval   = 2 * time.Second  // Обновление метрик каждые 2 секунды
	defaultReportInterval = 10 * time.Second // Отправка метрик каждые 10 секунд
	defaultServerAddress  = "localhost:8080"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := parseAgentFlags()
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	log.Printf("Starting metrics agent with config:")
	log.Printf("  Server address: %s", config.ServerAddress)
	log.Printf("  Poll interval: %v", config.PollInterval)
	log.Printf("  Report interval: %v", config.ReportInterval)

	// Создаем компоненты
	collector := agent.NewCollector()

	// Формируем полный URL для сервера
	serverURL := config.ServerAddress
	if !contains(serverURL, "http://") && !contains(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}

	sender := agent.NewSender(serverURL)

	// Контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	// Горутина для сбора метрик с настраиваемым интервалом
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Started metrics collection with interval: %v", config.PollInterval)

		ticker := time.NewTicker(config.PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping metrics collection...")
				return
			case <-ticker.C:
				collector.UpdateMetrics()
				gaugeCount, counterCount := collector.GetMetricsCount()
				log.Printf("Collected metrics: %d gauges, %d counters",
					gaugeCount, counterCount)
			}
		}
	}()

	// Горутина для отправки метрик с настраиваемым интервалом
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Started metrics reporting with interval: %v", config.ReportInterval)

		ticker := time.NewTicker(config.ReportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping metrics reporting...")
				return
			case <-ticker.C:
				// Получаем все метрики
				gauges := collector.GetGauges()
				counters := collector.GetCounters()

				if len(gauges) == 0 && len(counters) == 0 {
					log.Println("No metrics to send")
					continue
				}

				// Отправляем все метрики
				log.Printf("Sending metrics to %s", serverURL)
				if err := sender.SendAllMetrics(gauges, counters); err != nil {
					log.Printf("Failed to send metrics: %v", err)
				} else {
					log.Printf("Successfully sent all metrics")
				}
			}
		}
	}()

	// Ожидание сигнала завершения
	log.Println("Agent is running. Press Ctrl+C to stop.")
	<-ctx.Done()
	log.Println("Received shutdown signal...")

	// Graceful shutdown
	stop() // Останавливаем получение новых сигналов

	// Ждем завершения всех горутин с таймаутом
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Agent stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Println("Shutdown timeout, forcing exit")
	}

	return nil
}

// parseAgentFlags парсит флаги командной строки для агента
func parseAgentFlags() (*AgentConfig, error) {
	config := &AgentConfig{}

	var reportInterval, pollInterval int

	flag.StringVar(&config.ServerAddress, "a", defaultServerAddress, "HTTP server endpoint address")
	flag.IntVar(&reportInterval, "r", int(defaultReportInterval.Seconds()), "Report interval in seconds")
	flag.IntVar(&pollInterval, "p", int(defaultPollInterval.Seconds()), "Poll interval in seconds")

	// Парсим флаги
	flag.Parse()

	// Проверяем на неизвестные флаги
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown arguments: %v\n", flag.Args())
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Валидация значений
	if reportInterval <= 0 {
		fmt.Fprintf(os.Stderr, "Error: report interval must be positive, got %d\n", reportInterval)
		os.Exit(1)
	}

	if pollInterval <= 0 {
		fmt.Fprintf(os.Stderr, "Error: poll interval must be positive, got %d\n", pollInterval)
		os.Exit(1)
	}

	// Конвертируем в time.Duration
	config.ReportInterval = time.Duration(reportInterval) * time.Second
	config.PollInterval = time.Duration(pollInterval) * time.Second

	return config, nil
}

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s[:len(substr)] == substr ||
			(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
			indexOf(s, substr) != -1)
}

// indexOf находит индекс подстроки в строке
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
