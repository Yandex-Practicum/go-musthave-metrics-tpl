package handlers

import (
	"encoding/json"
	"net/http"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
)

func UpdatesJSONHandler(svc service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, m := range metrics {
			switch metricType(m.MType) {
			case typeGauge:
				if m.Value != nil {
					svc.UpdateGauge(r.Context(), m.ID, *m.Value)
				}
			case typeCounter:
				if m.Delta != nil {
					svc.UpdateCounter(r.Context(), m.ID, *m.Delta)
				}
			}
		}
		w.Header().Set("Contetn-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
