package main

import (
	"net/http"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/router"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"evgen3000/go-musthave-metrics-tpl.git/internal/config"
	log "evgen3000/go-musthave-metrics-tpl.git/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func runServer(host string, router *chi.Mux) {
	logger := log.GetLogger()
	logger.Info("server is running on", zap.String("host", host))
	err := http.ListenAndServe(host, router)
	if err != nil {
		logger.Fatal("Error", zap.String("Error", err.Error()))
	}
}

func main() {
	log.InitLogger()
	c := config.GetHost()
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)

	runServer(c.Value, r)
}

// curl -X POST http://localhost:8080/update/gauges/myGauge/3.14159 -H "Content-Type: text/plain"
// curl -X POST http://localhost:8080/update/counter/myGauge/5 -H "Content-Type: text/plain"
