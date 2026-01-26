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

func NewServer(h *handlers.Handlers, dbPingFunc func(context.Context) error) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/ping", h.PingHandler())
	// Добавьте другие роуты

	return &Server{router: r}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
