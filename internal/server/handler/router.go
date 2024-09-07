package handler

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/vova4o/yandexadv/internal/models"
)

// Router структура для роутера
type Router struct {
	Middl   Middlewarer
	mux     *gin.Engine
	Service Servicer
	server  *http.Server
	stopCh  chan struct{}
	mu      sync.Mutex
}

// Middlewarer интерфейс для middleware
type Middlewarer interface {
	GinZap() gin.HandlerFunc
	GunzipMiddleware() gin.HandlerFunc
	GzipMiddleware() gin.HandlerFunc
}

// Servicer интерфейс для сервиса
type Servicer interface {
	UpdateServ(metric models.Metric) error
	UpdateServJSON(metric *models.Metrics) error
	GetValueServ(metric models.Metrics) (string, error)
	GetValueServJSON(metric models.Metrics) (*models.Metrics, error)
	MetrixStatistic() (*template.Template, map[string]models.Metrics, error)
	PingDB() error
}

// New создание нового роутера
func New(s Servicer, middleware Middlewarer) *Router {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	return &Router{
		Middl:   middleware,
		mux:     router,
		Service: s,
		stopCh:  make(chan struct{}),
	}
}

// RegisterRoutes регистрация маршрутов
func (s *Router) RegisterRoutes() {

	s.mux.Use(s.Middl.GinZap())
	s.mux.Use(s.Middl.GunzipMiddleware())
	s.mux.Use(s.Middl.GzipMiddleware())

	s.mux.POST("/update/:type/:name/:value", s.UpdateMetricHandler)
	s.mux.GET("/value/:type/:name", s.GetValueHandler)
	s.mux.GET("/", s.StatisticPage)
	s.mux.POST("/update/", s.UpdateMetricHandlerJSON)
	s.mux.POST("/value/", s.GetValueHandlerJSON)
	s.mux.GET("/ping", s.PingHandler)
}

// StartServer запуск сервера
func (s *Router) StartServer(addr string) error {
	s.mu.Lock()
	// Создание http.Server с использованием Gin
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}
	s.mu.Unlock()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// Логирование ошибки, если сервер не смог запуститься
		log.Println("failed to start server", err)
		panic(err)
	}

	<-s.stopCh
	return nil
}

// StopServer остановка сервера
func (s *Router) StopServer(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.stopCh)
	// Остановка сервера с использованием контекста
	return s.server.Shutdown(ctx)
}
