package handlers

import (
	"fmt"
	"net/http"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
)

func ValueHandler(repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		switch metricType {
		case typeGauge:
			value, ok := repo.GetGauge(name)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%g", value)
		case typeCounter:
			value, ok := repo.GetCounter(name)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%d", value)
		default:
			http.Error(w, "unknown metric type", http.StatusNotFound)
		}
	}
}
