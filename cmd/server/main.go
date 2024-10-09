package main

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/handlers"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
)

func gracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Shutting down gracefully")
		os.Exit(0)
	}()
}

func runServer(host string, router *chi.Mux) {
	fmt.Println("Server is running on", host)
	err := http.ListenAndServe(host, router)
	if err != nil {
		panic(err)
	}
}

func main() {
	gracefulShutdown()
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")
	flag.Parse()
	env, isEnv := os.LookupEnv("ADDRESS")
	storage := storage.NewMemStorage()

	router := chi.NewRouter()
	handlers.HomeHandle(storage, router)
	handlers.UpdateHandler(storage, router)
	handlers.GetHandler(storage, router)

	if isEnv {
		runServer(env, router)
	} else {
		runServer(*hostFlag, router)
	}
}

// curl -X POST http://localhost:8080/update/gauges/myGauge/3.14159 -H "Content-Type: text/plain"
// curl -X POST http://localhost:8080/update/counter/myGauge/5 -H "Content-Type: text/plain"
