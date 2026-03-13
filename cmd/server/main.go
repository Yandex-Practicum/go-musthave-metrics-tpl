package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/config"
	handlers "github.com/kvsukharev/go-musthave-metrics-tpl/internal/handler"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/middleware_proj"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/server"
	"github.com/kvsukharev/go-musthave-metrics-tpl/internal/storage"
)

type ServerConfig struct {
	Address       string `env:"ADDRESS"`
	StoreInterval time.Duration
	FileStorage   string
	Restore       bool
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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

func (s *Server) Router() http.Handler {
	panic("unimplemented")
}

func NewServer(storage *MetricsStorage, config *ServerConfig) *Server {
	return &Server{
		storage: storage,
		config:  config,
	}
}

func main() {
	cfg, err := config.ParseFlags()
	if err := run(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	// Инициализация хранилища
	var store storage.Storage
	if cfg.DatabaseDSN != "" {
		dbStore, err := storage.NewPostgresStorage(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		log.Println("Using PostgreSQL storage")
		defer dbStore.Close()
	} else {
		store = storage.NewMemStorage()
	}

	// Создание обработчиков
	h := handlers.NewHandlers(store)

	// Инициализация роутера
	r := chi.NewRouter()

	// Применение middleware
	r.Use(
		middleware.Logger,
		middleware.Recoverer,
		middleware_proj.GzipMiddleware,
	)

	// Регистрация маршрутов
	h.RegisterRoutes(r)

	// Запуск сервера
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))

	// Инициализация хранилища
	var dbStorage *storage.PostgresStorage
	if cfg.DatabaseDSN != "" {
		dbStorage, err = storage.NewPostgresStorage(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Database init error: %v", err)
		}
		defer dbStorage.Close()
	}

	// Инициализация компонентов
	metricsStorage := handlers.NewMetricsStorage()
	handler := handlers.NewHandlers(metricsStorage)
	srv := server.NewServer(
		handler,
		func(ctx context.Context) error {
			return dbStorage.Ping(ctx)
		},
	)

	log.Printf("Starting server on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, srv); err != nil {
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
