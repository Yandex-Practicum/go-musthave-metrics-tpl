package main

import (
	"crypto/rsa"
	"os"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/agent"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/buildinfo"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/crypto"
	"github.com/rs/zerolog/log"
)

// Сведения о сборке. Значения по умолчанию можно перезаписать при компиляции
// через -ldflags "-X main.buildVersion=... -X main.buildDate=... -X main.buildCommit=...".
// Подробнее — в README проекта.
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	buildinfo.Print(os.Stdout, buildVersion, buildDate, buildCommit)

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("ошибка конфигурации")
	}

	baseURL := "http://" + cfg.Addr

	var publicKey *rsa.PublicKey
	if cfg.CryptoKey != "" {
		publicKey, err = crypto.LoadPublicKey(cfg.CryptoKey)
		if err != nil {
			log.Fatal().Err(err).Msg("ошибка загрузки публичного ключа")
		}
		log.Info().Str("path", cfg.CryptoKey).Msg("шифрование трафика включено")
	}

	sender := agent.NewSender(baseURL, cfg.HashKey, publicKey)
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
