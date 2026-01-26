package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/audit"
)

// TestableRun запускает сервер для тестирования с проверкой готовности
func TestableRun() error {
	// Используем существующую функцию run, но с улучшенной проверкой запуска
	config, err := parseServerFlags()
	if err != nil {
		return err
	}

	// Создаем зависимости
	storage := NewMetricsStorage()

	// Инициализация аудиторов
	var auditors []audit.Auditor
	if config.AuditFile != "" {
		auditors = append(auditors, audit.NewFileAuditor(config.AuditFile))
	}
	if config.AuditURL != "" {
		auditors = append(auditors, audit.NewHTTPAuditor(config.AuditURL))
	}

	server := NewServer(storage, config, auditors)

	if config.DatabaseDSN != "" {
		db, err := sql.Open("postgres", config.DatabaseDSN)
		if err != nil {
			// не выходим — логируем, но можно попытаться продолжить
		} else {
			// попробуем ping, чтобы убедиться в доступности
			if err := db.Ping(); err != nil {
				// оставляем db, чтобы последующая проверка могла показать проблему
			} else {
				// Connected to DB
			}
			server.db = db
		}
	}
	// Восстановление при старте (если включено)
	if config.FileStorage != "" && config.Restore {
		if err := storage.RestoreFromFile(config.FileStorage); err != nil {
			// Failed to restore from file
		} else {
			// Restored metrics from file
		}
	}

	// Фоновое периодическое сохранение или синхронная запись
	if config.FileStorage != "" {
		if config.StoreInterval == 0 {
			// Store interval = 0: synchronous writes enabled
		} else {
			// периодическое сохранение
			go func() {
				ticker := time.NewTicker(config.StoreInterval)
				defer ticker.Stop()
				for range ticker.C {
					if err := storage.SaveToFile(config.FileStorage); err != nil {
						// Failed to save metrics
					} else {
						// Saved metrics to file
					}
				}
			}()
		}
	}

	// Запускаем HTTP сервер в любом случае (server всегда используется)
	log.Printf("Starting metrics server on %s", config.Address)

	// Создаем сервер с контекстом для корректного завершения
	srv := &http.Server{
		Addr:    config.Address,
		Handler: server.Router(),
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Проверяем, что сервер запущен
	ready := make(chan struct{})
	go func() {
		for {
			conn, err := net.DialTimeout("tcp", config.Address, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				ready <- struct{}{}
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-ready:
		// Сервер готов
		return nil
	case <-time.After(5 * time.Second):
		// Timeout - сервер не готов
		return fmt.Errorf("server failed to start within timeout")
	}
}
