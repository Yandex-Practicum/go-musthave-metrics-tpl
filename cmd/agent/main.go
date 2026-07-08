package main

import (
	"os"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/agent"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/buildinfo"
	"github.com/rs/zerolog/log"
)

// Сведения о сборке. Задаются при компиляции через
// -ldflags "-X main.buildVersion=... -X main.buildDate=... -X main.buildCommit=...".
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	buildinfo.Print(os.Stdout, buildVersion, buildDate, buildCommit)

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("ошибка конфигурации")
	}

	baseURL := "http://" + cfg.Addr
	sender := agent.NewSender(baseURL, cfg.HashKey)
	store := agent.NewMetricsStore()

	go func() {
		ticker := time.NewTicker(cfg.PollInterval)
		for range ticker.C {
			store.UpdateGauges(agent.CollectGauges())
		}
	}()

	go func() {
		ticker := time.NewTicker(cfg.PollInterval)
		for range ticker.C {
			store.UpdatePSMetrics(agent.CollectPSUtilMetrics())
		}
	}()

	sem := make(chan struct{}, cfg.RateLimit)
	reportTicker := time.NewTicker(cfg.ReportInterval)

	for range reportTicker.C {
		all, ps := store.GetAll()

		sem <- struct{}{}
		go func(metrics []agent.GaugeMetric, count int64) {
			defer func() { <-sem }()
			if err := sender.SendBatch(metrics, count); err != nil {
				log.Error().Err(err).Msg("отправка метрик")
			}
		}(all, ps)
	}
}
