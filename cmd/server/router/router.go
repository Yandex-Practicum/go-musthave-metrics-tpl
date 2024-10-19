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

	chiRouter.Post("/update/", logger.HandlerLog(h.UpdateMetricHandlerJSON))
	chiRouter.Get("/value/", logger.HandlerLog(h.GetMetricHandlerJSON))

	chiRouter.Post("/update/{metricType}/{metricName}/{metricValue}", logger.HandlerLog(h.UpdateMetricHandlerText))
	chiRouter.Get("/value/{metricType}/{metricName}", logger.HandlerLog(h.GetMetricHandlerText))

	chiRouter.Get("/", logger.HandlerLog(h.HomeHandler))

	return chiRouter
}
