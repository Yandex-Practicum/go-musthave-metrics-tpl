// Package server собирает HTTP-сервер сбора метрик: настраивает роутер chi,
// подключает middleware (логирование, gzip, проверку подписи), регистрирует
// эндпоинты практического трека и обеспечивает graceful shutdown.
package server

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/audit"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/handlers"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/logger"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/middleware"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// Server — HTTP-сервер сбора метрик с зависимостями: адрес прослушивания,
// хранилище метрик, соединение с БД, ключ подписи, издатель аудита и флаг
// включения эндпоинтов профилирования pprof.
type Server struct {
	addr        string
	repo        storage.Repository
	db          *sql.DB
	hashKey     string
	auditPub    *audit.Publisher
	enablePprof bool
}

// New создаёт сервер с заданным адресом, хранилищем, соединением с БД,
// ключом HMAC-подписи (пустой — подпись отключена), издателем аудита и
// флагом enablePprof (регистрация /debug/pprof включается только когда true).
func New(addr string, repo storage.Repository, db *sql.DB, hashKey string, auditPub *audit.Publisher, enablePprof bool) *Server {
	return &Server{
		addr:        addr,
		repo:        repo,
		db:          db,
		hashKey:     hashKey,
		auditPub:    auditPub,
		enablePprof: enablePprof,
	}
}

// buildRouter собирает chi-роутер со всеми middleware и эндпоинтами сервера.
// Вынесен из Run отдельно, чтобы маршрутизацию можно было покрыть тестами
// без запуска блокирующего ListenAndServe.
func (s *Server) buildRouter() http.Handler {
	svc := service.NewMetricsService(s.repo)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Use(middleware.GzipMiddleware)
	// Эндпоинты pprof регистрируются только при явном включении:
	// они раскрывают внутренности процесса (стеки, heap) и не должны быть
	// доступны в production без ограничений. Включай в dev/staging при
	// профилировании (флаг --pprof / env PPROF).
	// go tool pprof http://<addr>/debug/pprof/heap
	if s.enablePprof {
		r.Mount("/debug", chimw.Profiler())
	}
	r.Post("/update/{type}/{name}/{value}", handlers.MetricsHandler(svc, s.auditPub))
	r.Post("/update/", handlers.UpdateJSONHandler(svc, s.auditPub))
	r.Get("/value/{type}/{name}", handlers.ValueHandler(svc))
	r.Post("/value/", handlers.ValueJSONHandler(svc))
	r.Get("/", handlers.ListHandler(svc))
	r.Get("/ping", handlers.PingHandler(s.db))
	r.Post("/updates/", handlers.UpdatesJSONHandler(svc, s.auditPub))
	if s.hashKey != "" {
		r.Use(middleware.HashCheckMiddleware(s.hashKey))
	}
	return r
}

// Run настраивает роутер и запускает HTTP-сервер, блокируясь до получения
// сигнала прерывания (os.Interrupt), после чего выполняет graceful shutdown.
func (s *Server) Run() error {
	r := s.buildRouter()

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
