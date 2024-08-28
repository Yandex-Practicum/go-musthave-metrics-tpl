package main

import (
	"log"

	"github.com/vova4o/yandexadv/internal/server/flags"
	"github.com/vova4o/yandexadv/internal/server/handler"
	"github.com/vova4o/yandexadv/internal/server/middleware"
	"github.com/vova4o/yandexadv/internal/server/service"
	"github.com/vova4o/yandexadv/internal/server/storage"
	"github.com/vova4o/yandexadv/package/logger"
)

func main() {
	config := flags.NewConfig()

	logger, err := logger.NewLogger("info", config.ServerLogFile)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	middle := middleware.New(logger)

	storage := storage.New()

	service := service.New(storage)

	router := handler.New(service, middle)
	router.RegisterRoutes()

	logger.Info("Starting server on " + config.ServerAddress)
	if err := router.StartServer(config.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
