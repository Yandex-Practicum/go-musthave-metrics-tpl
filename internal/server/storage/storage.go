package storage

import (
	"log"

	"github.com/vova4o/yandexadv/internal/models"
	"github.com/vova4o/yandexadv/internal/server/flags"
	"github.com/vova4o/yandexadv/package/logger"
	"go.uber.org/zap"
)

// Storage интерфейс для хранилища
type Storage interface {
	UpdateBatch(metrics []models.Metrics) error
	UpdateMetric(metric models.Metrics) error
	GetValue(metric models.Metrics) (*models.Metrics, error)
	MetrixStatistic() (map[string]models.Metrics, error)
	Ping() error
	Stop() error
}

// Init инициализация хранилища в зависимости от конфигурации
func Init(config *flags.Config, logger *logger.Logger) Storage {
	if config.FileStoragePath == "" && config.DBDSN == "" {
		logger.Error("No storage selected using default: MemoryStorage")
		return NewMemStorage()
	} else if config.DBDSN != "" {
		logger.Info("Selected storage: DB")
		DB, err := DBConnect(config, logger)
		if err != nil {
			logger.Error("Failed to connect to database: %v", zap.Error(err))
			log.Fatalf("Failed to connect to database: %v", err)
		}
		err = DB.CreateTables()
		if err != nil {
			logger.Error("Failed to create tables: %v", zap.Error(err))
			log.Fatalf("Failed to create tables: %v", err)
		}
		return DB
	} else {
		logger.Info("Selected storage: File")
		stor := NewFileStorage()
		StartFileStorageLogic(config, stor, logger)
		return stor
	}
}