package main

import (
	"flag"
	"log"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/server"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

func main() {
	addr := flag.String("a", ":8080", "адрес сервера (host:port)")
	flag.Parse()

	repo := storage.NewMemoryStorage()
	srv := server.New(*addr, repo)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
