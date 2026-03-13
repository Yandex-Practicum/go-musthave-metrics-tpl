package handlers

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/model"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	storage storage.Storage
}

func NewHandlers(storage storage.Storage) *Handlers {
	return &Handlers{storage: storage}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Post("/update", h.updateMetricJSONHandler)
	r.Post("/value", h.valueMetricJSONHandler)
	r.Post("/updates", h.BatchUpdateMetrics)
	r.Post("/update/*", h.updateHandler)
	r.Post("/update/{type}/{name}/{value}", h.updateHandlerChi)
	r.Get("/value/{type}/{name}", h.valueHandler)
	r.Get("/", h.rootHandler)
	r.Get("/ping", h.PingHandler())
}

type MetricsStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

// Close implements storage.Storage.
func (m *MetricsStorage) Close() error {
	panic("unimplemented")
}

// GetAllMetrics implements storage.Storage.
func (m *MetricsStorage) GetAllMetrics() (map[string]float64, map[string]int64) {
	panic("unimplemented")
}

// GetCounter implements storage.Storage.
func (m *MetricsStorage) GetCounter(name string) (int64, error) {
	panic("unimplemented")
}

// GetGauge implements storage.Storage.
func (m *MetricsStorage) GetGauge(name string) (float64, error) {
	panic("unimplemented")
}

// Ping implements storage.Storage.
func (m *MetricsStorage) Ping(ctx context.Context) error {
	panic("unimplemented")
}

// UpdateCounter implements storage.Storage.
func (m *MetricsStorage) UpdateCounter(name string, value int64) {
	panic("unimplemented")
}

// UpdateGauge implements storage.Storage.
func (m *MetricsStorage) UpdateGauge(name string, value float64) {
	panic("unimplemented")
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (h *Handlers) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.storage.Ping(r.Context()); err != nil {
			http.Error(w, "Database unavailable", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) valueMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string `json:"id"`
		MType string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	switch req.MType {
	case "gauge":
		value, err := h.storage.GetGauge(req.ID)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    req.ID,
			"type":  "gauge",
			"value": value,
		})

	case "counter":
		value, err := h.storage.GetCounter(req.ID)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    req.ID,
			"type":  "counter",
			"delta": value,
		})

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}

func decodeBody(r *http.Request, v interface{}) error {
	// Обработка gzip
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return err
		}
		defer gz.Close()
		return json.NewDecoder(gz).Decode(v)
	}
	return json.NewDecoder(r.Body).Decode(v)

}

func (h *Handlers) BatchUpdateMetrics(w http.ResponseWriter, r *http.Request) {
	var metrics []model.Metrics

	// Декодируем тело запроса
	if err := decodeBody(r, &metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем наличие метрик
	if len(metrics) == 0 {
		http.Error(w, "empty metrics batch", http.StatusBadRequest)
		return
	}

	// Выполняем пакетное обновление
	if err := h.storage.BatchUpdate(r.Context(), metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) updateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	var metric model.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			http.Error(w, "Missing value for gauge", http.StatusBadRequest)
			return
		}
		h.storage.UpdateGauge(metric.ID, *metric.Value)
	case "counter":
		if metric.Delta == nil {
			http.Error(w, "Missing delta for counter", http.StatusBadRequest)
			return
		}
		h.storage.UpdateCounter(metric.ID, *metric.Delta)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// updateHandler обрабатывает запросы на обновление метрик
func (h *Handlers) updateHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *Handlers) updateHandlerChi(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")
	h.updateMetric(w, metricType, metricName, metricValue)
}

func (h *Handlers) updateMetric(w http.ResponseWriter, metricType, metricName, metricValue string) {
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

func (h *Handlers) valueHandler(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprintf(w, "%g", value)

	case "counter":
		value, err := h.storage.GetCounter(metricName)
		if err != nil {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "%d", value)

	default:
		http.Error(w, "Unknown metric type. Use 'gauge' or 'counter'", http.StatusBadRequest)
	}
}

func (h *Handlers) rootHandler(w http.ResponseWriter, r *http.Request) {
	gauges, counters := h.storage.GetAllMetrics()

	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>Metrics Server</title>
    <style>
        /* ... (ваши стили остаются без изменений) ... */
    </style>
</head>
<body>
    <div class="container">
        <h1>Metrics Server Dashboard</h1>
        
        <h2>Gauges <span class="count">({{len .Gauges}})</span></h2>
        <table>
            <tr><th>Name</th><th>Value</th></tr>
            {{range $name, $value := .Gauges}}
            <tr><td><strong>{{$name}}</strong></td><td>{{printf "%.6f" $value}}</td></tr>
            {{else}}
            <tr><td colspan="2" style="text-align: center; color: #666;">No gauges available</td></tr>
            {{end}}
        </table>
        
        <h2>Counters <span class="count">({{len .Counters}})</span></h2>
        <table>
            <tr><th>Name</th><th>Value</th></tr>
            {{range $name, $value := .Counters}}
            <tr><td><strong>{{$name}}</strong></td><td>{{$value}}</td></tr>
            {{else}}
            <tr><td colspan="2" style="text-align: center; color: #666;">No counters available</td></tr>
            {{end}}
        </table>
        
        <div style="margin-top: 30px; padding: 15px; background-color: #e7f3ff; border-left: 4px solid #2196F3;">
            <h3>API Endpoints:</h3>
            <ul>
                <li><code>POST /update/{type}/{name}/{value}</code> - Update metric</li>
                <li><code>GET /value/{type}/{name}</code> - Get metric value</li>
                <li><code>GET /</code> - This dashboard</li>
                <li><code>GET /ping</code> - Check database connection</li>
            </ul>
        </div>
    </div>
</body>
</html>`

	t, err := template.New("metrics").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	data := struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := t.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
