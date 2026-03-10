package main

import (
	"log"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/server"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

func main() {
	repo := storage.NewMemoryStorage()
	srv := server.New(":5050", repo)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
