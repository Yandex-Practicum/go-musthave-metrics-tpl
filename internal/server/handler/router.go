package handler

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/vova4o/yandexadv/internal/models"
)

// Router структура для роутера
type Router struct {
	mux     *gin.Engine
	Service Servicer
}

// Servicer интерфейс для сервиса
type Servicer interface {
	UpdateServ(metric models.Metric) error
	GetValueServ(metric models.Metric) (string, error)
	MetrixStatistic() (*template.Template, map[string]interface{}, error)
}

// New создание нового роутера
func New(s Servicer) *Router {
	return &Router{
		mux:     gin.Default(),
		Service: s,
	}
}

// RegisterRoutes регистрация маршрутов
func (s *Router) RegisterRoutes() {
	s.mux.POST("/update/:type/:name/:value", s.UpdateMetricHandler)
	s.mux.GET("/value/:type/:name", s.GetValueHandler)
	s.mux.GET("/", s.StatisticPage)
}

// StartServer запуск сервера
func (s *Router) StartServer(addr string) error {
	// Запуск сервера с использованием Gin
	if err := s.mux.Run(addr); err != nil {
		return err
	}
	return nil
}
