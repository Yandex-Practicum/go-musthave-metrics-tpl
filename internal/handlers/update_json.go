package handlers

import (
	"encoding/json"
	"net/http"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

func UpdateJSONHandler(repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch m.MType {
		case typeGauge:
			if m.Value == nil {
				http.Error(w, "value is required", http.StatusBadRequest)
				return
			}
			repo.UpdateGauge(m.ID, *m.Value)
		case typeCounter:
			if m.Delta == nil {
				http.Error(w, "delta is required", http.StatusBadRequest)
				return
			}
			repo.UpdateCounter(m.ID, *m.Delta)
		default:
			http.Error(w, "unknown metric type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(m)
	}
}
