package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vova4o/yandexadv/internal/server/flags"
	"github.com/vova4o/yandexadv/internal/server/handler"
	"github.com/vova4o/yandexadv/internal/server/middleware"
	"github.com/vova4o/yandexadv/internal/server/service"
	"github.com/vova4o/yandexadv/internal/server/storage"
	"github.com/vova4o/yandexadv/package/logger"
	"go.uber.org/zap"
)

func main() {
	config := flags.NewConfig()

	logger, err := logger.NewLogger("info", config.ServerLogFile)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	middle := middleware.New(logger)

	stor := storage.New()

	DB, err := storage.DBConnect(config)
	if err != nil {
		logger.Error("Failed to connect to database: %v", zap.Error(err))
	}
	stor.DB = DB

	service := service.New(stor)

	router := handler.New(service, middle)
	router.RegisterRoutes()

	// We can add IF statment later if we have to.
	storage.StartFileStorageLogic(config, stor, logger)

	// Создание канала для получения сигналов завершения работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	go func() {
		logger.Info("Starting server on " + config.ServerAddress)
		if err := router.StartServer(config.ServerAddress); err != nil {
			logger.Error("Failed to start server: %v", zap.Error(err))
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидание сигнала завершения работы
	<-stop

	// Создание контекста с тайм-аутом для завершения работы сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stor.SaveMemStorageToFile()
	stor.CloseFile()

	// stor.SavetoDB()
	// stor.CloseDB()

	// Логирование завершения работы сервера
	logger.Info("Shutting down server...")

	// Завершение работы сервера
	if err := router.StopServer(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", zap.Error(err))
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
