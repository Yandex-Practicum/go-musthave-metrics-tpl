package handlers

import (
	"encoding/json"
	"net/http"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

func ValueJSONHandler(repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch m.MType {
		case typeGauge:
			value, ok := repo.GetGauge(m.ID)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			m.Value = &value
		case typeCounter:
			value, ok := repo.GetCounter(m.ID)
			if !ok {
				http.Error(w, "metric not found", http.StatusNotFound)
				return
			}
			m.Delta = &value
		default:
			http.Error(w, "unknown metric type", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}
