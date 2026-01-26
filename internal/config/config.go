package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Key           string        `env:"KEY"`
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	FileStorage   string        `env:"FILE_STORAGE"`
	Restore       bool          `env:"RESTORE"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
	RateLimit     int           `env:"RATE_LIMIT"`
	AuditFile     string        `env:"AUDIT_FILE"`
	AuditURL      string        `env:"AUDIT_URL"`
}

type Config struct {
	Key           string
	Address       string
	DatabaseDSN   string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	AuditFile     string
	AuditURL      string
}

const (
	defaultAddress       = "localhost:8080"
	defaultStoreInterval = 300 * time.Second
	defaultFileStorage   = "metrics-db.json"
	defaultRestore       = false
	defaultDatabaseDSN   = "localhost"
)

func loadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{
		Address:       defaultAddress,
		StoreInterval: defaultStoreInterval,
		FileStorage:   defaultFileStorage,
		Restore:       defaultRestore,
		DatabaseDSN:   defaultDatabaseDSN,
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

	flag.StringVar(&cfg.Key, "k", os.Getenv("KEY"), "Secret key for HMAC")
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

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		cfg.AuditURL = flagAuditURL
	} else if envAuditURL != "" {
		cfg.AuditURL = envAuditURL
	}

	return cfg, nil
}

func ParseFlags() (*ServerConfig, error) {
	cfg := &ServerConfig{
		Address:   "localhost:8080",
		RateLimit: 5,
	}

	flag.IntVar(&cfg.RateLimit, "RATE_LIMIT", 5, "Max concurrent requests")
	flag.StringVar(&cfg.Key, "k", os.Getenv("KEY"), "Secret key for HMAC")
	flag.StringVar(&cfg.Address, "a", cfg.Address, "HTTP server endpoint address")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database connection string")
	flag.Parse()

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DatabaseDSN = envDSN
	}

	if flag.NArg() > 0 {
		return nil, fmt.Errorf("unknown arguments: %v", flag.Args())
	}

	return cfg, nil
}

func FlagsLoad() *Config {
	cfg := &Config{
		Address:       "localhost:8080",
		StoreInterval: 300 * time.Second,
		StoreFile:     "/tmp/metrics-db.json",
	}

	flag.StringVar(&cfg.Key, "k", os.Getenv("KEY"), "Secret key for HMAC")
	flag.StringVar(&cfg.Address, "a", cfg.Address, "Server address")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "PostgreSQL DSN")
	flag.StringVar(&cfg.StoreFile, "f", cfg.StoreFile, "File storage path")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore metrics from file")
	flag.Parse()

	// Environment variables
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DatabaseDSN = envDSN
	}
	if envFile := os.Getenv("STORE_FILE"); envFile != "" {
		cfg.StoreFile = envFile
	}

	return cfg
}
