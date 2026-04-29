package server

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/handlers"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/logger"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/middleware"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type Server struct {
	addr    string
	repo    storage.Repository
	db      *sql.DB
	hashKey string
	server  *http.Server
}

func New(addr string, repo storage.Repository, db *sql.DB, hashKey string) *Server {
	return &Server{
		addr:    addr,
		repo:    repo,
		db:      db,
		hashKey: hashKey,
	}
}

func (s *Server) Run() error {
	svc := service.NewMetricsService(s.repo)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.HashCheckMiddleware(s.hashKey))
	r.Post("/update/{type}/{name}/{value}", handlers.MetricsHandler(svc))
	r.Post("/update/", handlers.UpdateJSONHandler(svc))
	r.Get("/value/{type}/{name}", handlers.ValueHandler(svc))
	r.Post("/value/", handlers.ValueJSONHandler(svc))
	r.Get("/", handlers.ListHandler(svc))
	r.Get("/ping", handlers.PingHandler(s.db))
	r.Post("/updates/", handlers.UpdatesJSONHandler(svc))

	srv := &http.Server{
		Addr:         s.addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Ошибка при запуска сервера")
		}
	}()
	log.Info().Str("addr", s.addr).Msg("сервер запущен")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info().Msg("Выключене сервера ... ")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Сервер был отключен")
	}
	log.Info().Msg("Сервер завершил работу корректно")
	return nil
}
