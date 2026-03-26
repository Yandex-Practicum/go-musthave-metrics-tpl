package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"
)

type agentConfig struct {
	Addr           string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func parseConfig() agentConfig {
	addr := flag.String("a", "localhost:5050", "адрес сервера (host:port)")
	pollInterval := flag.Duration("p", 2*time.Second, "интервал сбора метрик")
	reportInterval := flag.Duration("r", 10*time.Second, "интервал отправки метрик на сервер")
	flag.Parse()

	//перезапись
	if v := os.Getenv("ADDRESS"); v != "" {
		*addr = v
	}

	if v := os.Getenv("REPORT_INTERVAL"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("неверное значение REPORT_INTERVAL: %v", err)
		}
		*reportInterval = time.Duration(sec) * time.Second
	}

	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("неверное значение POLL_INTERVAL : %v", err)
		}
		*pollInterval = time.Duration(sec) * time.Second
	}
	//
	return agentConfig{
		Addr:           *addr,
		PollInterval:   *pollInterval,
		ReportInterval: *reportInterval,
	}

}
