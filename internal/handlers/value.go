package handlers

import (
	"fmt"
	"net/http"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
	"github.com/go-chi/chi/v5"
)

func ValueHandler(srv service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		switch metricType(mType) {
		case typeGauge:
			value, ok := srv.GetGauge(r.Context(), name)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%g", value)
		case typeCounter:
			value, ok := srv.GetCounter(r.Context(), name)
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
