package main

import (
	"sync"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/agent"
	"github.com/rs/zerolog/log"
)

func main() {
	var (
		mu         sync.Mutex
		lastGauges []agent.GaugeMetric
		lastPS     []agent.GaugeMetric
		pollCount  int64
	)

	cfg := parseConfig()

	baseURL := "http://" + cfg.Addr
	sender := agent.NewSender(baseURL, cfg.HashKey)

	go func() {
		ticker := time.NewTicker(cfg.PollInterval)
		for range ticker.C {
			gauges := agent.CollectGauges()
			mu.Lock()
			lastGauges = gauges
			pollCount++
			mu.Unlock()
		}
	}()

	go func() {
		ticker := time.NewTicker(cfg.PollInterval)
		for range ticker.C {
			ps := agent.CollectPSUtilMetrics()
			mu.Lock()
			lastPS = ps
			mu.Unlock()
		}
	}()

	sem := make(chan struct{}, cfg.RateLimit)
	reportTicker := time.NewTicker(cfg.ReportInterval)

	for range reportTicker.C {
		mu.Lock()
		all := append(lastGauges, lastPS...)
		ps := pollCount
		mu.Unlock()

		sem <- struct{}{}
		go func(metrics []agent.GaugeMetric, count int64) {
			defer func() { <-sem }()
			if err := sender.SendBatch(metrics, count); err != nil {
				log.Error().Err(err).Msg("отправка метрик")
			}
		}(all, ps)
	}
}
