package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/audit"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/buildinfo"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/logger"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/server"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/rs/zerolog/log"
)

// Сведения о сборке. Значения по умолчанию можно перезаписать при компиляции
// через -ldflags "-X main.buildVersion=... -X main.buildDate=... -X main.buildCommit=...".
// Подробнее — в README проекта.
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	buildinfo.Print(os.Stdout, buildVersion, buildDate, buildCommit)

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("ошибка конфигурации")
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal().Err(err).Msg("ошибка инициализации логгера")
	}

	var repo storage.Repository
	var db *sql.DB

	if cfg.DatabaseDSN != "" {
		var err error
		db, err = storage.NewPostgresDB(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal().Err(err).Msg("ошибка при подключения к БД")
		}
		repo = storage.NewPostgresStorege(db)
	} else {
		memRepo := storage.NewMemoryStorage()

		// загрузка метрик из файла при старте
		if cfg.Restore && cfg.FilePath != "" {
			if err := storage.LoadFromFile(memRepo, cfg.FilePath); err != nil {
				log.Error().Err(err).Msg("Не удалось загрузить метрик")
			}
		}
		// периодическое сохранение на диск
		if cfg.FilePath != "" && cfg.StoreInterval > 0 {
			go storage.RunSaver(memRepo, cfg.FilePath, time.Duration(cfg.StoreInterval)*time.Second)
		}
		//синхронная запись — передаём filePath в сервер
		if cfg.FilePath != "" && cfg.StoreInterval == 0 {
			memRepo.SetSyncFile(cfg.FilePath)
		}
		repo = memRepo
	}

	// Аудит запросов (паттерн «Наблюдатель»): приёмники подключаются,
	// только если задан соответствующий параметр конфигурации.
	auditPub := audit.NewPublisher()
	if cfg.AuditFile != "" {
		fs, err := audit.NewFileSink(cfg.AuditFile)
		if err != nil {
			log.Fatal().Err(err).Msg("ошибка открытия файла аудита")
		}
		auditPub.Subscribe(fs)
		log.Info().Str("file", cfg.AuditFile).Msg("аудит в файл включён")
	}
	if cfg.AuditURL != "" {
		auditPub.Subscribe(audit.NewHTTPSink(cfg.AuditURL))
		log.Info().Str("url", cfg.AuditURL).Msg("аудит по сети включён")
	}
	defer func() {
		if err := auditPub.Close(); err != nil {
			log.Error().Err(err).Msg("ошибка закрытия приёмников аудита")
		}
	}()

	srv := server.New(cfg.Addr, repo, db, cfg.HashKey, auditPub, cfg.EnablePprof)
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("ошибка запуска сервера")
	}

}
