package main

import (
	"flag"
	"log"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/agent"
)

func main() {
	addr := flag.String("a", "localhost:5050", "адрес сервера (host:port)")
	pollInterval := flag.Duration("p", 2*time.Second, "интервал сбора метрик")
	reportInterval := flag.Duration("r", 10*time.Second, "интервал отправки метрик на сервер")
	flag.Parse()

	baseURL := "http://" + *addr
	sender := agent.NewSender(baseURL)

	var lastGauges []agent.GaugeMetric
	var pollsSinceReport int64
	reportTick := *reportInterval

	for {
		lastGauges = agent.CollectGauges()
		pollsSinceReport++

		for elapsed := time.Duration(0); elapsed < reportTick; elapsed += *pollInterval {
			time.Sleep(*pollInterval)
			lastGauges = agent.CollectGauges()
			pollsSinceReport++
		}
		if err := sender.SendAll(lastGauges, pollsSinceReport); err != nil {
			log.Printf("отправка метрик: %v", err)
			continue
		}
		log.Printf("метрики отправлены (polls=%d)", pollsSinceReport)
		pollsSinceReport = 0
	}
}
