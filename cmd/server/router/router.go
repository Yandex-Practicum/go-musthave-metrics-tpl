package router

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/handlers"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"

	"github.com/go-chi/chi/v5"
)

func SetupRouter(storage *storage.MemStorage) *chi.Mux {
    h := handlers.NewHandler()
	chiRouter := chi.NewRouter()

	chiRouter.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMetricHandler(storage))
	chiRouter.Get("/value/{metricType}/{metricName}", h.GetMetricHandler(storage))
	chiRouter.Get("/", h.HomeHandler(storage)) // Переносим HomeHandler в handlers

	return chiRouter
}
