package main

import (
	"log"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/agent"
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
			if err := sender.SendAll(lastGauges, pollsSinceReport); err != nil {
				log.Printf("отправка метрик: %v", err)
				continue
			}
			log.Printf("метрики отправлены (polls=%d)", pollsSinceReport)
			pollsSinceReport = 0
		}
	}
}
