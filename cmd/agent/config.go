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
			log.Fatal().Err(err).Msg("неверное значение REPORT_INTERVAL")
		}
		*reportInterval = time.Duration(sec) * time.Second
	}

	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal().Err(err).Msg("неверное значение POLL_INTERVAL")
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
