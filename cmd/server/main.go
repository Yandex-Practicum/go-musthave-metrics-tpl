package main

import (
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/handlers"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)
func main() {
	storage := storage.NewMemStorage()

	router := chi.NewRouter()

	handlers.HomeHandle(storage, router)
	handlers.UpdateHandler(storage, router)
	handlers.GetHandler(storage, router)

	fmt.Println("Server is running on http://localhost:8080")
	err := http.ListenAndServe("localhost:8080", router)
	if err != nil {
		panic(err)
	}
}

// curl -X POST http://localhost:8080/update/gauges/myGauge/3.14159 -H "Content-Type: text/plain"
// curl -X POST http://localhost:8080/update/counter/myGauge/5 -H "Content-Type: text/plain"
