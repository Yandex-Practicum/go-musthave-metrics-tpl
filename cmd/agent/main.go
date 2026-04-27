package main

import (
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/agent"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := parseConfig()

	baseURL := "http://" + cfg.Addr
	sender := agent.NewSender(baseURL)

	var lastGauges []agent.GaugeMetric
	var pollsSinceReport int64

	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)

	for {
		select {
		case <-pollTicker.C:
			lastGauges = agent.CollectGauges()
			pollsSinceReport++
		case <-reportTicker.C:
			if err := sender.SendBatch(lastGauges, pollsSinceReport); err != nil {
				log.Error().Err(err).Msg("отправка метрик")
				continue
			}
			log.Info().Int64("polls", pollsSinceReport).Msg("метрики отправлены")
			pollsSinceReport = 0
		}
	}
}
