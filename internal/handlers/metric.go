package handlers

import (
	"net/http"
	"strconv"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
	"github.com/go-chi/chi/v5"
)

type metricType string

const (
	typeGauge   metricType = "gauge"
	typeCounter metricType = "counter"
)

func MetricsHandler(srv service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		mType := chi.URLParam(r, "type")
		name := chi.URLParam(r, "name")
		valueStr := chi.URLParam(r, "value")

		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch metricType(mType) {
		case typeGauge:
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			srv.UpdateGauge(r.Context(), name, value)
		case typeCounter:
			delta, err := strconv.ParseInt(valueStr, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			srv.UpdateCounter(r.Context(), name, delta)
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
