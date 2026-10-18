package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	h "yandex-go-musthave-metrics-tpl/internal/handler"
	mw "yandex-go-musthave-metrics-tpl/internal/handler/middleware"
	ms "yandex-go-musthave-metrics-tpl/internal/repository/memstorage"
	s "yandex-go-musthave-metrics-tpl/internal/service/server"

	"github.com/go-chi/chi/v5"
)

func main() {
	addr := flag.String("a", "localhost:8080", "Address to listen server on")
	flag.Parse()
	if val, ok := os.LookupEnv("ADDRESS"); ok {
		*addr = val
	}
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)
	repository := ms.NewMemStorage()
	service := s.NewMetricService(repository)
	handler := h.NewMetricHandler(service)
	middleware := mw.NewMiddleware(logger)
	r := chi.NewRouter()
	router := r.With(middleware.WithLogger)
	router.Post("/update/{type}/{name}/{value}", handler.Update)
	router.Get("/value/{type}/{name}", handler.Get)
	router.Get("/", handler.GetAll)
	fmt.Println("Listening on", *addr)
	if err := http.ListenAndServe(*addr, router); err != nil {
		panic(err)
	}
}
