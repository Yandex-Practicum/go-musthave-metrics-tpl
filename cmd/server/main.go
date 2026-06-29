package main

import (
	"database/sql"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/logger"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/server"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/rs/zerolog/log"
)

func main() {
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

	srv := server.New(cfg.Addr, repo, db, cfg.HashKey)
	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("ошибка запуска сервера")
	}

}
