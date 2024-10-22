package main

import (
	"net/http"
	"time"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/router"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"evgen3000/go-musthave-metrics-tpl.git/internal/config"
	log "evgen3000/go-musthave-metrics-tpl.git/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func runServer(config *config.ServerConfig, router *chi.Mux) {
	logger := log.GetLogger()
	logger.Info("server is running on", zap.String("host", config.Host))
	err := http.ListenAndServe(config.Host, router)
	if err != nil {
		logger.Fatal("Error", zap.String("Error", err.Error()))
	}
}

func main() {
	log.InitLogger()
	c := config.GetServerConfig()
	s := storage.NewMemStorage(storage.MemStorageConfig{
		StoreInterval:   c.StoreInterval,
		FileStoragePath: c.FilePath,
		Restore:         c.Restore,
	})

	r := router.SetupRouter(s)

	ticker := time.NewTicker(time.Duration(c.StoreInterval) * time.Second)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if err := s.SaveData(c.FilePath); err != nil {
				log.GetLogger().Fatal("Can't to save data", zap.Error(err))
			} else {
				log.GetLogger().Info("Saved data")
			}
		}
	}()

	runServer(c, r)
}
