package handlers

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {}

func NewHandler() *Handler {
	return &Handler{}
}

func (h Handler) HomeHandler(storage *storage.MemStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		body := "<h4>Gauges</h4>"
		for gaugeName, value := range storage.GetAllGauges() {
			body += gaugeName + ": " + strconv.FormatFloat(value, 'f', -1, 64) + "</br>"
		}
		body += "<h4>Counters</h4>"

		for counterName, value := range storage.GetAllCounters() {
			body += counterName + ": " + strconv.FormatInt(value, 10) + "</br>"
		}
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")

		_, err := rw.Write([]byte(body))
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}

func (h Handler) UpdateMetricHandler(storage *storage.MemStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		switch metricType {
		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, "Bad request", http.StatusBadRequest)
				return
			}
			storage.IncrementCounter(metricName, value)
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, "Bad request", http.StatusBadRequest)
				return
			}
			storage.SetGauge(metricName, value)
		default:
			http.Error(rw, "Bad request", http.StatusBadRequest)
			return
		}

		rw.WriteHeader(http.StatusOK)
	}
}

func (h Handler) GetMetricHandler(storage *storage.MemStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType != "gauge" && metricType != "counter" {
			http.Error(rw, "Invlid metric type", http.StatusBadRequest)
			return
		}

		switch metricType {
		case "gauge":
			value, exists := storage.GetGauge(metricName)
			if !exists {
				http.Error(rw, "Metric not found", http.StatusNotFound)
				return
			}
			rw.Header().Set("Content-type", "text/plain")
			_, err := rw.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))
			if err != nil {
				log.Printf("Write failed: %v", err)
			}
		case "counter":
			value, exists := storage.GetCounter(metricName)
			if !exists {
				http.Error(rw, "Metric not found", http.StatusNotFound)
				return
			}
			rw.Header().Set("Content-type", "text/plain")
			_, err := rw.Write([]byte(strconv.FormatInt(value, 10)))
			if err != nil {
				log.Printf("Write failed: %v", err)
			}
		default:
			http.Error(rw, "Invalid metric type", http.StatusBadRequest)
			return
		}

	}
}
