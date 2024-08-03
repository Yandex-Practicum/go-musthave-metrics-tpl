package main

import (
	"log"

	"github.com/vova4o/yandexadv/internal/server/flags"
	"github.com/vova4o/yandexadv/internal/server/handler"
	"github.com/vova4o/yandexadv/internal/server/service"
	"github.com/vova4o/yandexadv/internal/server/storage"
)

func main() {
	config := flags.NewConfig()

	storage := storage.New()
	
	service := service.New(storage)

	router := handler.New(service)
	router.RegisterRoutes()

	log.Println("Starting server on "+ config.ServerAddress)
	if err := router.StartServer(config.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
