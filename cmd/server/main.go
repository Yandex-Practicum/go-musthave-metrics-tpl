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


func main() {
	storage := &MemStorage{gauges: make(map[string]float64), counters: make(map[string]int64)}
	fmt.Println("Server is running on http://localhost:8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", updateHandler(storage))
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func badRequest(w http.ResponseWriter) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func updateHandler(storage *MemStorage) http.HandlerFunc {
	// Возвращаем анонимную функцию (обработчик)
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.Split(req.URL.Path, "/")
		fmt.Println(path)
		// update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
		if len(path) != 5 {
			http.Error(w, "Invalid URL format", http.StatusNotFound)
			return
		}
		if req.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Invalid data format", http.StatusNotFound)
		}

		metricType, metricName, metricValue := path[2], path[3], path[4]

		switch metricType {
		case "counters":
			pathValue, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				badRequest(w)
				return
			}
			storage.counters[metricName] += pathValue
			w.WriteHeader(http.StatusOK)
		case "gauges":
			pathValue, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				badRequest(w)
				return
			}
			storage.gauges[metricName] = pathValue
		default:
			badRequest(w)
		}
	}
}

// curl -X POST http://localhost:8080/update/gauges/myGauge/3.14159 -H "Content-Type: text/plain"
// curl -X POST http://localhost:8080/update/counters/myGauge/5 -H "Content-Type: text/plain"