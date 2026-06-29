package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/audit"
	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
	"github.com/rs/zerolog/log"
)

func UpdatesJSONHandler(svc service.MetricsService, auditPub *audit.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		names := make([]string, 0, len(metrics))
		for _, m := range metrics {
			switch metricType(m.MType) {
			case typeGauge:
				if m.Value != nil {
					if err := svc.UpdateGauge(r.Context(), m.ID, *m.Value); err != nil {
						log.Error().Err(err).Msg("ошибка обновления gauge")
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					names = append(names, m.ID)
				}
			case typeCounter:
				if m.Delta != nil {
					if err := svc.UpdateCounter(r.Context(), m.ID, *m.Delta); err != nil {
						log.Error().Err(err).Msg("ошибка обновления gauge")
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					names = append(names, m.ID)
				}
			}
		}
		publishAudit(auditPub, r, names)
		w.Header().Set("Contetn-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
