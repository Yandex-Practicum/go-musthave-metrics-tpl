package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type serverConfig struct {
	Addr          string
	LogLevel      string
	StoreInterval int
	FilePath      string
	Restore       bool
	DatabaseDSN   string
	HashKey       string
	AuditFile     string
	AuditURL      string
	EnablePprof   bool
}

func parseConfig() (serverConfig, error) {
	addr := flag.String("a", ":8080", "адрес сервера")
	logLevel := flag.String("l", "info", "уровень логирования")
	storeInterval := flag.Int("i", 300, "интервал сохранения на диск")
	filePath := flag.String("f", "metrics.json", "путь до файла")
	restore := flag.Bool("r", true, "загружать при старте")
	databaseDSN := flag.String("d", "", "строка подключения к PostgreSQL")
	hashKey := flag.String("k", "", "ключ для подписи SHA256")
	auditFile := flag.String("audit-file", "", "путь к файлу логов аудита (пусто — аудит в файл отключён)")
	auditURL := flag.String("audit-url", "", "URL приёмника логов аудита (пусто — аудит по сети отключён)")
	enablePprof := flag.Bool("pprof", false, "включить эндпоинты /debug/pprof (только для dev/staging)")
	flag.Parse()

	if v, ok := os.LookupEnv("ADDRESS"); ok {
		*addr = v
	}
	if v, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		sec, err := strconv.Atoi(v)
		if err != nil {
			return serverConfig{}, fmt.Errorf("неверное значение STORE_INTERVAL: %w", err)
		}
		*storeInterval = sec
	}
	if v, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		*filePath = v
	}
	if v, ok := os.LookupEnv("RESTORE"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return serverConfig{}, fmt.Errorf("неверное значение RESTORE: %w", err)
		}
		*restore = b
	}
	if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
		*databaseDSN = v
	}
	if v, ok := os.LookupEnv("KEY"); ok {
		*hashKey = v
	}
	if v, ok := os.LookupEnv("AUDIT_FILE"); ok {
		*auditFile = v
	}
	if v, ok := os.LookupEnv("AUDIT_URL"); ok {
		*auditURL = v
	}
	if v, ok := os.LookupEnv("PPROF"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return serverConfig{}, fmt.Errorf("неверное значение PPROF: %w", err)
		}
		*enablePprof = b
	}

	return serverConfig{
		Addr:          *addr,
		LogLevel:      *logLevel,
		StoreInterval: *storeInterval,
		FilePath:      *filePath,
		Restore:       *restore,
		DatabaseDSN:   *databaseDSN,
		HashKey:       *hashKey,
		AuditFile:     *auditFile,
		AuditURL:      *auditURL,
		EnablePprof:   *enablePprof,
	}, nil
}
