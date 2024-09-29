package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

var storage = MemStorage{gauges: make(map[string]float64), counters: make(map[string]int64)}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func badRequest(w http.ResponseWriter) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func updateHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Only POST method are allowed", http.StatusMethodNotAllowed)
	}

	path := strings.Split(req.URL.Path, "/")
	fmt.Println(path)

	// update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>

	if len(path) != 4 && req.Header.Get("Content-type") != "text/plain" {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
	}

	switch path[1] {
	case "counters":
		pathName := path[2]
		pathValue, err := strconv.ParseInt(path[3], 10, 64)
		if err != nil {
			badRequest(w)
		}
		storage.counters[pathName] += pathValue
	case "gauges":
		pathName := path[2]
		pathValue, err := strconv.ParseFloat(path[3], 64)
		if err != nil {
			badRequest(w)
		}
		storage.gauges[pathName] = pathValue
	default:
		badRequest(w)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update", updateHandler)
	return http.ListenAndServe(":8080", mux)
}

