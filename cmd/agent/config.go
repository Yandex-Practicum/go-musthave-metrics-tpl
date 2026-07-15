package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type agentConfig struct {
	Addr           string
	PollInterval   time.Duration
	ReportInterval time.Duration
	HashKey        string
	RateLimit      int
	CryptoKey      string
}

func parseConfig() (agentConfig, error) {
	addr := flag.String("a", "localhost:8080", "адрес сервера (host:port)")
	pollInterval := flag.Int("p", 2, "интервал сбора метрик (сек)")
	reportInterval := flag.Int("r", 10, "интервал отправки метрик на сервер (сек)")
	hashKey := flag.String("k", "", "ключ для подписи SHA256")
	rateLimit := flag.Int("l", 1, "количество одновременных запросов")
	cryptoKey := flag.String("crypto-key", "", "путь до файла с публичным RSA-ключом (пусто — шифрование отключено)")

	flag.Parse()

	if v, ok := os.LookupEnv("ADDRESS"); ok {
		*addr = v
	}

	if v, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		sec, err := strconv.Atoi(v)
		if err != nil {
			return agentConfig{}, fmt.Errorf("неверное значение REPORT_INTERVAL: %w", err)
		}
		*reportInterval = sec
	}

	if v, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		sec, err := strconv.Atoi(v)
		if err != nil {
			return agentConfig{}, fmt.Errorf("неверное значение POLL_INTERVAL: %w", err)
		}
		*pollInterval = sec
	}

	if v, ok := os.LookupEnv("RATE_LIMIT"); ok {
		rl, err := strconv.Atoi(v)
		if err != nil {
			return agentConfig{}, fmt.Errorf("неверное значение RATE_LIMIT: %w", err)
		}
		*rateLimit = rl
	}

	if v, ok := os.LookupEnv("KEY"); ok {
		*hashKey = v
	}

	if v, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		*cryptoKey = v
	}

	return agentConfig{
		Addr:           *addr,
		PollInterval:   time.Duration(*pollInterval) * time.Second,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
		HashKey:        *hashKey,
		RateLimit:      *rateLimit,
		CryptoKey:      *cryptoKey,
	}, nil
}
