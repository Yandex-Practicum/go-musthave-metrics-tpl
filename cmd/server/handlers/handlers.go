package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
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
	storage *storage.MemStorage
}

func NewHandler(storage *storage.MemStorage) *Handler {
	return &Handler{storage}
}

func (h *Handler) HomeHandler(rw http.ResponseWriter, _ *http.Request) {
	var body strings.Builder
	body.WriteString("<h4>Gauges</h4>")
	for gaugeName, value := range h.storage.GetAllGauges() {
		body.WriteString(gaugeName + ": " + strconv.FormatFloat(value, 'f', -1, 64) + "</br>")
	}
	body.WriteString("<h4>Counters</h4>")

	for counterName, value := range h.storage.GetAllCounters() {
		body.WriteString(counterName + ": " + strconv.FormatInt(value, 10) + "</br>")
	}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := rw.Write([]byte(body.String()))
	if err != nil {
		log.Printf("Write failed: %v", err)
	}
}

func (h *Handler) UpdateMetricHandler(rw http.ResponseWriter, r *http.Request) {
	var body Metrics

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	switch body.MType {
	case MetricTypeCounter:
		h.storage.IncrementCounter(body.ID, *body.Delta)
	case MetricTypeGauge:
		h.storage.SetGauge(body.ID, *body.Value)
	default:
		http.Error(rw, "Bad request", http.StatusBadRequest)
	}

	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	var body Metrics

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if body.MType != MetricTypeGauge && body.MType != MetricTypeCounter {
		http.Error(rw, "Invalid metric type", http.StatusBadRequest)
	}

	if body.MType == MetricTypeGauge {
		value, exists := h.storage.GetGauge(body.ID)
		if !exists {
			http.Error(rw, "Metric not found", http.StatusNotFound)
		}
		rw.Header().Set("Content-type", "application/json")
		jsonBody, _ := json.Marshal(Metrics{ID: body.ID, MType: body.MType, Value: &value})
		_, err := rw.Write(jsonBody)
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	} else {
		value, exists := h.storage.GetCounter(body.ID)
		if !exists {
			http.Error(rw, "Metric not found", http.StatusNotFound)
		}
		rw.Header().Set("Content-type", "application/json")
		jsonBody, _ := json.Marshal(Metrics{ID: body.ID, MType: body.MType, Delta: &value})
		_, err := rw.Write(jsonBody)
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}
