package handlers

import (
	"encoding/json"
	"net/http"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
)

func UpdateJSONHandler(srv service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch metricType(m.MType) {
		case typeGauge:
			if m.Value == nil {
				http.Error(w, "value is required", http.StatusBadRequest)
				return
			}
			if err := srv.UpdateGauge(r.Context(), m.ID, *m.Value); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case typeCounter:
			if m.Delta == nil {
				http.Error(w, "delta is required", http.StatusBadRequest)
				return
			}
			if err := srv.UpdateCounter(r.Context(), m.ID, *m.Delta); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "unknown metric type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(m)
	}
}
