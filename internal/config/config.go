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
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	FileStorage   string        `env:"FILE_STORAGE"`
	Restore       bool          `env:"RESTORE"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}

type Config struct {
	Address       string
	DatabaseDSN   string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
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
		flagAddress  string
		flagInterval int
		flagFile     string
		flagRestore  bool
	)

	flag.StringVar(&flagAddress, "a", "", "Server address")
	flag.IntVar(&flagInterval, "i", -1, "Store interval in seconds (0 = sync write)")
	flag.StringVar(&flagFile, "f", "", "File path for storage")
	flag.BoolVar(&flagRestore, "r", false, "Restore from storage file on start")

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

	return cfg, nil
}

func ParseFlags() (*ServerConfig, error) {
	cfg := &ServerConfig{
		Address: "localhost:8080",
	}

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

func MustLoad() *Config {
	cfg := &Config{
		Address:       "localhost:8080",
		StoreInterval: 300 * time.Second,
		StoreFile:     "/tmp/metrics-db.json",
	}

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
