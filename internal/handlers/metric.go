package handlers

import (
	"net/http"
	"strconv"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
)

const (
	pathPrefixUpdate = "/update/"
	typeGauge        = "gauge"
	typeCounter      = "counter"
)

func MetricsHandler(repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		valueStr := chi.URLParam(r, "value")

		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch metricType {
		case typeGauge:
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			repo.UpdateGauge(name, value)
		case typeCounter:
			delta, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			repo.UpdateCounter(name, delta)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
