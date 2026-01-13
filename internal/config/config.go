package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Address       string
	StoreInterval time.Duration
	FileStorage   string
	Restore       bool
	AuditFile     string
	AuditURL      string
}

const (
	defaultAddress       = "localhost:8080"
	defaultStoreInterval = 300 * time.Second
	defaultFileStorage   = "metrics-db.json"
	defaultRestore       = false
)

func loadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{
		Address:       defaultAddress,
		StoreInterval: defaultStoreInterval,
		FileStorage:   defaultFileStorage,
		Restore:       defaultRestore,
	}

	// Загрузка из env vars
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	var (
		flagAddress   string
		flagInterval  int
		flagFile      string
		flagRestore   bool
		flagAuditFile string
		flagAuditURL  string
	)

	flag.StringVar(&flagAddress, "a", "", "Server address")
	flag.IntVar(&flagInterval, "i", -1, "Store interval in seconds (0 = sync write)")
	flag.StringVar(&flagFile, "f", "", "File path for storage")
	flag.BoolVar(&flagRestore, "r", false, "Restore from storage file on start")
	flag.StringVar(&flagAuditFile, "audit-file", "", "Path to audit log file")
	flag.StringVar(&flagAuditURL, "audit-url", "", "URL for audit logs")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr == "" && flagAddress != "" {
		cfg.Address = flagAddress
	}

	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval == "" {
		if flagInterval >= 0 {
			cfg.StoreInterval = time.Duration(flagInterval) * time.Second
		}
	} else {
		if val, err := strconv.Atoi(envInterval); err == nil {
			cfg.StoreInterval = time.Duration(val) * time.Second
		}
	}

	if envFile := os.Getenv("FILE_STORAGE_PATH"); envFile == "" && flagFile != "" {
		cfg.FileStorage = flagFile
	} else if envFile != "" {
		cfg.FileStorage = envFile
	}

	if envRestore := os.Getenv("RESTORE"); envRestore == "" {
		cfg.Restore = flagRestore
	} else {
		r := strings.ToLower(envRestore)
		cfg.Restore = r == "true" || r == "1"
	}

	// Обработка флагов аудита
	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile == "" && flagAuditFile != "" {
		cfg.AuditFile = flagAuditFile
	} else if envAuditFile != "" {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL == "" && flagAuditURL != "" {
		cfg.AuditURL = flagAuditURL
	} else if envAuditURL != "" {
		cfg.AuditURL = envAuditURL
	}

	return cfg, nil
}
