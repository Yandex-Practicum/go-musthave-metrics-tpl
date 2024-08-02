package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/vova4o/yandexadv/internal/models"
)

// Router структура для роутера
type Router struct {
    mux *http.ServeMux
	Service Servicer
}

// Servicer интерфейс для сервиса
type Servicer interface {
	Update(metricType, metricName string, metricValue interface{}) error
}

// New создание нового роутера
func New(s Servicer) *Router {
    return &Router{
        mux: http.NewServeMux(),
		Service: s,
    }
}

// RegisterRoutes регистрация маршрутов
func (s *Router) RegisterRoutes() {
    // r.mux.HandleFunc("/update/", r.updateHandler)
    s.mux.HandleFunc("/update/", s.UpdateMetricHandler())
}

// StartServer запуск сервера
func (s *Router) StartServer(addr string) error {
    return http.ListenAndServe(addr, s.mux)
}

// UpdateMetricHandler обработчик для обновления метрики
func (s *Router) UpdateMetricHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, req *http.Request) {
        if req.Method != http.MethodPost {
            http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
            return
        }

        parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/update/"), "/")
        if len(parts) != 3 {
            http.Error(w, "Invalid request format", http.StatusNotFound)
            return
        }

        metricType, metricName, metricValue := parts[0], parts[1], parts[2]

        err := s.Service.Update(metricType, metricName, metricValue)
        if err != nil {
            if httpErr, ok := err.(*models.HTTPError); ok {
                log.Printf("Error: %v", httpErr.Message)
                http.Error(w, httpErr.Message, httpErr.Status)
                return
            }
            log.Printf("Internal server error: %v", err)
            http.Error(w, "internal server error", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    }
}

