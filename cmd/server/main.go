package main

import (
	"log"
	"net/http"
	"time"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/router"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage/filemanager"
	"evgen3000/go-musthave-metrics-tpl.git/internal/config/server"
	httpLogger "evgen3000/go-musthave-metrics-tpl.git/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func runServer(config *server.Config, router *chi.Mux) {
	logger := httpLogger.InitLogger()
	logger.Info("server is running on", zap.String("host", config.Host))
	err := http.ListenAndServe(config.Host, router)
	if err != nil {
		logger.Fatal("Error", zap.String("Error", err.Error()))
	}
}

func main() {
	c := server.GetServerConfig()
	fs := filemanager.FileManager{}
	s := storage.NewMemStorage(storage.MemStorageConfig{
		StoreInterval:   c.StoreInterval,
		FileStoragePath: c.FilePath,
		Restore:         c.Restore,
	}, &fs)

	r := router.SetupRouter(s)

	ticker := time.NewTicker(c.StoreInterval)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if err := fs.SaveData(c.FilePath, s); err != nil {
				log.Fatal("Can't to save data", zap.Error(err))
			} else {
				log.Println("Saved data")
			}
		}
	}()

	runServer(c, r)
}
