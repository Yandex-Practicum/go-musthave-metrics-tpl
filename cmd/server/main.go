package main

import (
	"log"

	"github.com/vova4o/yandexadv/internal/server/handler"
	"github.com/vova4o/yandexadv/internal/server/service"
	"github.com/vova4o/yandexadv/internal/server/storage"
)


func main() {
	storage := storage.New()
	
	service := service.New(storage)

	router := handler.New(service)
	router.RegisterRoutes()

	log.Println("Starting server on :8080")
	if err := router.StartServer(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
