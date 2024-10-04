package main

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/handlers"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"flag"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)
func main() {
	hostFlag := flag.String("a", "localhost:8080", "Host IP address and port.")

	storage := storage.NewMemStorage()

	router := chi.NewRouter()

	handlers.HomeHandle(storage, router)
	handlers.UpdateHandler(storage, router)
	handlers.GetHandler(storage, router)

	flag.Parse()
	fmt.Println("Server is running on", *hostFlag)
	err := http.ListenAndServe(*hostFlag, router)
	if err != nil {
		panic(err)
	}
}

// curl -X POST http://localhost:8080/update/gauges/myGauge/3.14159 -H "Content-Type: text/plain"
// curl -X POST http://localhost:8080/update/counter/myGauge/5 -H "Content-Type: text/plain"
