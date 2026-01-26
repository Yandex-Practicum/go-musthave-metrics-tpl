// Package handlers provides HTTP handlers for metrics operations.
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/storage"

	"github.com/go-chi/chi/v5"
)

// MetricHandlers handles HTTP requests for metrics operations.
type MetricHandlers struct {
	// storage is the storage backend for metrics
	storage storage.Storage
}

// NewMetricHandlers creates and returns a new MetricHandlers instance.
func NewMetricHandlers(storage storage.Storage) *MetricHandlers {
	return &MetricHandlers{storage: storage}
}

// updateHandler handles requests to update metrics via path parameters.
// It expects URL parameters: type, name, and value.
func (h *MetricHandlers) updateHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(metricName, value)
		log.Printf("Updated gauge %s = %.6f", metricName, value)

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(metricName, value)
		log.Printf("Updated counter %s (added %d)", metricName, value)

	default:
		http.Error(w, "Unknown metric type. Use 'gauge' or 'counter'", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK\n")
}

// valueHandler handles requests to retrieve metric values via path parameters.
// It expects URL parameters: type and name.
func (h *MetricHandlers) valueHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch metricType {
	case "gauge":
		value, err := h.storage.GetGauge(metricName)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%g", value)

	case "counter":
		value, err := h.storage.GetCounter(metricName)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d", value)

	default:
		http.Error(w, "Unknown metric type. Use 'gauge' or 'counter'", http.StatusBadRequest)
	}
}
