package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ServerConfig struct {
	Address string
}

type MetricsStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

type Server struct {
	storage *MetricsStorage
	config  *ServerConfig
}

func NewServer(storage *MetricsStorage, config *ServerConfig) *Server {
	return &Server{
		storage: storage,
		config:  config,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Post("/update/*", s.updateHandler)
	r.Post("/update/{type}/{name}/{value}", s.updateHandlerChi)
	r.Get("/value/{type}/{name}", s.valueHandler)
	r.Get("/", s.rootHandler)

	return r
}

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/update/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		http.Error(w, "Invalid URL format. Expected: /update/{type}/{name}/{value}",
			http.StatusBadRequest)
		return
	}

	s.updateMetric(w, parts[0], parts[1], parts[2])
}

func (s *Server) updateHandlerChi(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	s.updateMetric(w, metricType, metricName, metricValue)
}

func (s *Server) updateMetric(w http.ResponseWriter, metricType, metricName, metricValue string) {
	s.storage.mu.Lock()
	defer s.storage.mu.Unlock()

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		s.storage.gauges[metricName] = value
		log.Printf("Updated gauge %s = %.6f", metricName, value)

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		s.storage.counters[metricName] += value
		log.Printf("Updated counter %s = %d (added %d)", metricName, s.storage.counters[metricName], value)

	default:
		http.Error(w, "Unknown metric type. Use 'gauge' or 'counter'",
			http.StatusBadRequest)
		return
	}

	responseText := "OK\n"
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(responseText)))
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, responseText)
}

func (s *Server) valueHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	s.storage.mu.RLock()
	defer s.storage.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch metricType {
	case "gauge":
		value, exists := s.storage.gauges[metricName]
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%g", value)

	case "counter":
		value, exists := s.storage.counters[metricName]
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d", value)

	default:
		http.Error(w, "Unknown metric type. Use 'gauge' or 'counter'", http.StatusBadRequest)
	}
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	// Создаем копии для безопасной работы с шаблоном
	gaugesCopy := make(map[string]float64)
	countersCopy := make(map[string]int64)

	// Блокируем только на время копирования
	s.storage.mu.RLock()
	for k, v := range s.storage.gauges {
		gaugesCopy[k] = v
	}
	for k, v := range s.storage.counters {
		countersCopy[k] = v
	}
	s.storage.mu.RUnlock()

	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>Metrics Server</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 40px; 
            background-color: #f5f5f5; 
        }
        .container {
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        table { 
            border-collapse: collapse; 
            width: 100%; 
            margin-bottom: 20px; 
        }
        th, td { 
            border: 1px solid #ddd; 
            padding: 12px; 
            text-align: left; 
        }
        th { 
            background-color: #4CAF50; 
            color: white;
        }
        tr:nth-child(even) {
            background-color: #f2f2f2;
        }
        h1 { 
            color: #333; 
            text-align: center;
        }
        h2 { 
            color: #4CAF50; 
            border-bottom: 2px solid #4CAF50;
            padding-bottom: 10px;
        }
        .count {
            color: #666;
            font-size: 0.9em;
        }
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
		Gauges:   gaugesCopy,
		Counters: countersCopy,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := t.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := parseServerFlags()
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	// Создаем зависимости
	storage := NewMetricsStorage()
	server := NewServer(storage, config)

	log.Printf("Starting metrics server on %s", config.Address)
	if err := http.ListenAndServe(config.Address, server.Router()); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

func parseServerFlags() (*ServerConfig, error) {
	config := &ServerConfig{}

	flag.StringVar(&config.Address, "a", "localhost:8080", "HTTP server endpoint address")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown arguments: %v\n", flag.Args())
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		return nil, fmt.Errorf("unknown arguments provided")
	}

	return config, nil
}
