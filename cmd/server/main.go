package main

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/router"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"evgen3000/go-musthave-metrics-tpl.git/internal/config"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func runServer(host string, router *chi.Mux) {
	fmt.Println("Server is running on", host)
	err := http.ListenAndServe(host, router)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	c := config.GetHost()
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)

	runServer(c.Value, r)
}

// curl -X POST http://localhost:8080/update/gauges/myGauge/3.14159 -H "Content-Type: text/plain"
// curl -X POST http://localhost:8080/update/counter/myGauge/5 -H "Content-Type: text/plain"
