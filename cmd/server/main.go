package main

import (
	"flag"
	"log"
	"os"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/logger"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/server"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

func main() {
	addr := flag.String("a", ":8080", "адрес сервера (host:port)")
	logLevel := flag.String("l", "info", "уровень логирования")
	flag.Parse()

	if v := os.Getenv("ADDRESS"); v != "" {
		*addr = v
	}

	if err := logger.Initialize(*logLevel); err != nil {
		log.Fatal(err)
	}

	repo := storage.NewMemoryStorage()
	srv := server.New(*addr, repo)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
