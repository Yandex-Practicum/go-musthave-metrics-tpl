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

	chiRouter.Route("/update", func(r chi.Router) {
		r.Post("/", logger.HandlerLog(h.UpdateMetricHandlerJSON))
		r.Post("/{metricType}/{metricName}/{metricValue}", logger.HandlerLog(h.UpdateMetricHandlerText))
	})

	chiRouter.Route("/value", func(r chi.Router) {
		r.Get("/", logger.HandlerLog(h.GetMetricHandlerJSON))
		r.Get("/{metricType}/{metricName}", logger.HandlerLog(h.GetMetricHandlerText))
	})

	chiRouter.Get("/", logger.HandlerLog(h.HomeHandler))

	return chiRouter
}
