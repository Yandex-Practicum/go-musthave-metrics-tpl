package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	handlers "github.com/kvsukharev/go-musthave-metrics-tpl/internal/handler"
)

type Server struct {
	router *chi.Mux
}

func NewServer(h *handlers.MetricHandlers, dbPingFunc func(context.Context) error) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Регистрация маршрутов для обработки метрик через path parameters
	r.Post("/update/{type}/{name}/{value}", h.UpdateHandler)
	r.Get("/value/{type}/{name}", h.ValueHandler)
	r.Get("/", h.RootHandler)
	r.Get("/ping", h.PingHandler)

	// Дополнительные маршруты для тестов
	r.Post("/update/counter/{name}/{value}", h.UpdateHandler)
	r.Post("/update/gauge/{name}/{value}", h.UpdateHandler)
	r.Get("/value/counter/{name}", h.ValueHandler)
	r.Get("/value/gauge/{name}", h.ValueHandler)

	// Маршруты с trailing slash для тестов
	r.Post("/update/counter/{name}/{value}/", h.UpdateHandler)
	r.Post("/update/gauge/{name}/{value}/", h.UpdateHandler)
	r.Get("/value/counter/{name}/", h.ValueHandler)
	r.Get("/value/gauge/{name}/", h.ValueHandler)

	// Маршруты для всех типов с trailing slash
	r.Post("/update/{type}/{name}/{value}/", h.UpdateHandler)

	return &Server{router: r}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
