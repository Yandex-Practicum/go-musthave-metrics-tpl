package server

import (
	"log"
	"net/http"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/handlers"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/logger"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/middleware"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	addr   string
	repo   storage.Repository
	server *http.Server
}

func New(addr string, repo storage.Repository) *Server {
	return &Server{
		addr: addr,
		repo: repo,
	}
}

func (s *Server) Run() error {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(middleware.GzipMiddleware)
	r.Post("/update/{type}/{name}/{value}", handlers.MetricsHandler(s.repo))
	r.Post("/update/", handlers.UpdateJSONHandler(s.repo))
	r.Get("/value/{type}/{name}", handlers.ValueHandler(s.repo))
	r.Post("/value/", handlers.ValueJSONHandler(s.repo))
	r.Get("/", handlers.ListHandler(s.repo))

	s.server = &http.Server{Addr: s.addr, Handler: r}
	log.Printf("сервер запущен на %s", s.addr)
	return s.server.ListenAndServe()
}
