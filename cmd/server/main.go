package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	h "yandex-go-musthave-metrics-tpl/internal/handler"
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
	repository := ms.NewMemStorage()
	service := s.NewMetricService(repository)
	handler := h.NewMetricHandler(service)
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handler.Update)
	r.Get("/value/{type}/{name}", handler.Get)
	r.Get("/", handler.GetAll)
	fmt.Println("Listening on", *addr)
	if err := http.ListenAndServe(*addr, r); err != nil {
		panic(err)
	}
}
