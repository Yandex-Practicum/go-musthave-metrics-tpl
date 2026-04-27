package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

type agentConfig struct {
	Addr           string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func parseConfig() agentConfig {
	addr := flag.String("a", "localhost:8080", "адрес сервера (host:port)")
	pollInterval := flag.Int("p", 2, "интервал сбора метрик (сек)")
	reportInterval := flag.Int("r", 10, "интервал отправки метрик на сервер (сек)")
	flag.Parse()

	if v := os.Getenv("ADDRESS"); v != "" {
		*addr = v
	}

	if v := os.Getenv("REPORT_INTERVAL"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal().Err(err).Msg("неверное значение REPORT_INTERVAL")
		}
		*reportInterval = sec
	}

	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal().Err(err).Msg("неверное значение POLL_INTERVAL")
		}
		*pollInterval = sec
	}

	return agentConfig{
		Addr:           *addr,
		PollInterval:   time.Duration(*pollInterval) * time.Second,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
	}
}
