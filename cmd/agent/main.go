package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/agent"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/logger"

	"github.com/go-chi/chi/v5"
)

// RootConfig – верхний уровень с вложенным agent_config
type RootConfig struct {
	AgentConfig AgentConfig `yaml:"agent_config"`
}

// AgentConfig с тегами yaml и env
type AgentConfig struct {
	server_address string        `yaml:"server_adress" env:"ADDRESS"` // Обращаем внимание: env тег использует точное имя переменной
	PollInterval   time.Duration `yaml:"poll_interval"`               // интервал в time.Duration, парсим отдельно
	ReportInterval time.Duration `yaml:"report_interval"`             // как выше
}

const (
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultServerAddress  = "localhost:8080"
	configPath            = "internal/config/agent.yaml"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	log := logger.GetLogger()

	cfg, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := applyEnv(cfg); err != nil {
		return fmt.Errorf("apply env: %w", err)
	}

	if err := parseFlags(cfg); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	log.Info().
		Str("Starting metrics agent with config:", "").
		Str("Server address: %s", cfg.server_address).
		Dur("Poll interval: %v", cfg.PollInterval).
		Dur("Report interval: %v", cfg.ReportInterval)

	client := &http.Client{Timeout: 10 * time.Second}
	collector := agent.NewCollector(100, client, "http://localhost:8080")

	serverURL := cfg.server_address
	if len(serverURL) < 7 || (serverURL[:7] != "http://" && serverURL[:8] != "https://") {
		serverURL = "http://" + serverURL
	}

	sender := agent.NewSender(serverURL)

	// Router и middleware с логированием
	r := chi.NewRouter()
	r.Use(logger.Middleware)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Dur("Started metrics collection with interval: %v", cfg.PollInterval)
		ticker := time.NewTicker(cfg.PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping metrics collection...")
				return
			case <-ticker.C:
				collector.UpdateMetrics()
				gCount, cCount := collector.GetMetricsCount()
				log.Info().Msgf("Collected metrics: %d gauges, %d counters", gCount, cCount)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msgf("Started metrics reporting with interval: %v", cfg.ReportInterval)
		ticker := time.NewTicker(cfg.ReportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Stopping metrics reporting...")
				return
			case <-ticker.C:
				gauges := collector.GetGauges()
				counters := collector.GetCounters()
				if len(gauges) == 0 && len(counters) == 0 {
					log.Info().Msg("No metrics to send")
					continue
				}
				log.Info().Str("Sending metrics to %s", serverURL)
				if err := sender.SendAllMetrics(gauges, counters); err != nil {
					log.Info().Msgf("Failed to send metrics: %v", err)
				} else {
					log.Info().Msg("Successfully sent all metrics")
				}
			}
		}
	}()

	log.Info().Msg("Agent is running. Press Ctrl+C to stop.")

	<-ctx.Done()
	log.Info().Msg("Received shutdown signal...")

	stop()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("Agent stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Info().Msg("Shutdown timeout, forcing exit")
	}

	return nil
}

// loadConfig читает YAML, задаёт дефолты, парсит в структуру
func loadConfig(path string) (*AgentConfig, error) {
	rootCfg := &RootConfig{
		AgentConfig: AgentConfig{
			server_address: defaultServerAddress,
			PollInterval:   defaultPollInterval,
			ReportInterval: defaultReportInterval,
		},
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Config file %q not found, using defaults and env variables", path)
	} else {
		if err := yaml.Unmarshal(data, rootCfg); err != nil {
			return nil, fmt.Errorf("unmarshal yaml: %w", err)
		}
	}

	return &rootCfg.AgentConfig, nil
}

// applyEnv проверяет переменные окружения и если они есть — перекрывает параметры
func applyEnv(cfg *AgentConfig) error {
	// Переменная окружения ADDRESS
	if addr := os.Getenv("ADDRESS"); addr != "" {
		cfg.server_address = addr
	}

	// Переменные интервалов интервалов в секундах — парсим из строк
	if pollStr := os.Getenv("POLL_INTERVAL"); pollStr != "" {
		sec, err := strconv.Atoi(pollStr)
		if err != nil {
			return fmt.Errorf("invalid POLL_INTERVAL: %w", err)
		}
		cfg.PollInterval = time.Duration(sec) * time.Second
	}

	if reportStr := os.Getenv("REPORT_INTERVAL"); reportStr != "" {
		sec, err := strconv.Atoi(reportStr)
		if err != nil {
			return fmt.Errorf("invalid REPORT_INTERVAL: %w", err)
		}
		cfg.ReportInterval = time.Duration(sec) * time.Second
	}

	return nil
}

// parseFlags применяет параметры из флагов, только если соответствующая ENV не задана (приоритет env выше)
func parseFlags(cfg *AgentConfig) error {
	var (
		flagAddress        string
		flagPollInterval   int
		flagReportInterval int
	)

	flag.StringVar(&flagAddress, "a", "", "HTTP server endpoint address")
	flag.IntVar(&flagPollInterval, "p", 0, "Poll interval in seconds")
	flag.IntVar(&flagReportInterval, "r", 0, "Report interval in seconds")

	flag.Parse()

	// Применяем флаги, если переменные окружения не заданы
	if os.Getenv("ADDRESS") == "" && flagAddress != "" {
		cfg.server_address = flagAddress
	}

	if os.Getenv("POLL_INTERVAL") == "" && flagPollInterval > 0 {
		cfg.PollInterval = time.Duration(flagPollInterval) * time.Second
	}

	if os.Getenv("REPORT_INTERVAL") == "" && flagReportInterval > 0 {
		cfg.ReportInterval = time.Duration(flagReportInterval) * time.Second
	}

	return nil
}
