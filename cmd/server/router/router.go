package router

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/handlers"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"evgen3000/go-musthave-metrics-tpl.git/internal/logger"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(storage *storage.MemStorage) *chi.Mux {
	h := handlers.NewHandler(storage)
	chiRouter := chi.NewRouter()

	chiRouter.Post("/update/", logger.HandlerLog(h.UpdateMetricHandler))
	chiRouter.Get("/value/", logger.HandlerLog(h.GetMetricHandler))
	chiRouter.Get("/", logger.HandlerLog(h.HomeHandler))

	return chiRouter
}
