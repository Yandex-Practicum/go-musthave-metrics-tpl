package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/audit"
	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
	"github.com/rs/zerolog/log"
)

func UpdateJSONHandler(srv service.MetricsService, auditPub *audit.Publisher) http.HandlerFunc {
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
				log.Error().Err(err).Msg("ошибка обновления gauge")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		case typeCounter:
			if m.Delta == nil {
				http.Error(w, "delta is required", http.StatusBadRequest)
				return
			}
			if err := srv.UpdateCounter(r.Context(), m.ID, *m.Delta); err != nil {
				log.Error().Err(err).Msg("ошибка обновления counter")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "unknown metric type", http.StatusBadRequest)
			return
		}

		publishAudit(auditPub, r, []string{m.ID})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(m)
	}
}
