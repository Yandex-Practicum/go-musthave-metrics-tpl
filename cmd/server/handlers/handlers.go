package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"github.com/go-chi/chi/v5"
)

const (
	MetricTypeCounter = "counter"
	MetricTypeGauge   = "gauge"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type Handler struct {
	Storage *storage.MemStorage
}

func NewHandler(storage *storage.MemStorage) *Handler {
	return &Handler{storage}
}

func (h *Handler) HomeHandler(rw http.ResponseWriter, _ *http.Request) {
	var body strings.Builder
	body.WriteString("<h4>Gauges</h4>")
	for gaugeName, value := range h.Storage.GetAllGauges() {
		body.WriteString(gaugeName + ": " + strconv.FormatFloat(value, 'f', -1, 64) + "</br>")
	}
	body.WriteString("<h4>Counters</h4>")

	for counterName, value := range h.Storage.GetAllCounters() {
		body.WriteString(counterName + ": " + strconv.FormatInt(value, 10) + "</br>")
	}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := rw.Write([]byte(body.String()))
	if err != nil {
		http.Error(rw, "Write failed: %v", http.StatusBadRequest)
	}
}

func (h *Handler) UpdateMetricHandlerJSON(rw http.ResponseWriter, r *http.Request) {
	var body Metrics
	rw.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	switch body.MType {
	case MetricTypeCounter:
		h.Storage.IncrementCounter(body.ID, *body.Delta)
		value, _ := h.Storage.GetCounter(body.ID)

		jsonBody, err := json.Marshal(Metrics{body.ID, body.MType, &value, nil})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		_, err = rw.Write(jsonBody)
		rw.WriteHeader(http.StatusOK)
		return
	case MetricTypeGauge:
		h.Storage.SetGauge(body.ID, *body.Value)
		value, _ := h.Storage.GetGauge(body.ID)

		jsonBody, err := json.Marshal(Metrics{body.ID, body.MType, nil, &value})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
		_, err = rw.Write(jsonBody)
		return
	default:
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}
}

func (h *Handler) UpdateMetricHandlerText(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	switch metricType {
	case MetricTypeCounter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(rw, "Bad request", http.StatusBadRequest)
		}
		h.Storage.IncrementCounter(metricName, value)
	case MetricTypeGauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(rw, "Bad request", http.StatusBadRequest)
		}
		h.Storage.SetGauge(metricName, value)
	default:
		http.Error(rw, "Bad request", http.StatusBadRequest)
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricHandlerJSON(rw http.ResponseWriter, r *http.Request) {
	var body Metrics
	rw.Header().Set("Content-Type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if body.MType != MetricTypeGauge && body.MType != MetricTypeCounter {
		http.Error(rw, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if body.MType == MetricTypeGauge {
		value, exists := h.Storage.GetGauge(body.ID)

		if !exists {
			http.Error(rw, "Metric not found", http.StatusNotFound)
			return
		}

		jsonBody, _ := json.Marshal(Metrics{ID: body.ID, MType: body.MType, Value: &value})
		_, err := rw.Write(jsonBody)
		if err != nil {
			http.Error(rw, "Write failed", http.StatusBadRequest)
			return
		}
	} else {
		value, exists := h.Storage.GetCounter(body.ID)
		if !exists {
			http.Error(rw, "Metric not found", http.StatusNotFound)
			return
		}
		jsonBody, err := json.Marshal(Metrics{ID: body.ID, MType: body.MType, Delta: &value})
		if err != nil {
			http.Error(rw, "Json write failed:", http.StatusBadRequest)
			return
		}

		_, err = rw.Write(jsonBody)
		if err != nil {
			http.Error(rw, "Write failed", http.StatusBadRequest)
			return
		}
	}
}

func (h *Handler) GetMetricHandlerText(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if metricType != MetricTypeGauge && metricType != MetricTypeCounter {
		http.Error(rw, "Invalid metric type", http.StatusBadRequest)
	}

	if metricType == MetricTypeGauge {
		value, exists := h.Storage.GetGauge(metricName)
		if !exists {
			http.Error(rw, "Metric not found", http.StatusNotFound)
		}
		rw.Header().Set("Content-Type", "text/plain")
		_, err := rw.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
		if err != nil {
			http.Error(rw, "Write failed", http.StatusBadRequest)
		}
	} else {
		value, exists := h.Storage.GetCounter(metricName)
		if !exists {
			http.Error(rw, "Metric not found", http.StatusNotFound)
		}
		rw.Header().Set("Content-Type", "text/plain")
		_, err := rw.Write([]byte(strconv.FormatInt(value, 10)))
		if err != nil {
			http.Error(rw, "Write failed", http.StatusBadRequest)
		}
	}
}
