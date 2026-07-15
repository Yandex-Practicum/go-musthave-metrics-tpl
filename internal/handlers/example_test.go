package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/audit"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/handlers"
	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
)

// newTestServer поднимает тестовый HTTP-сервер с in-memory хранилищем и
// зарегистрированными эндпоинтами практического трека. Аудит отключён
// (пустой Publisher).
func newTestServer() *httptest.Server {
	svc := service.NewMetricsService(storage.NewMemoryStorage())
	pub := audit.NewPublisher()

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handlers.MetricsHandler(svc, pub))
	r.Get("/value/{type}/{name}", handlers.ValueHandler(svc))
	r.Post("/update/", handlers.UpdateJSONHandler(svc, pub))
	r.Post("/value/", handlers.ValueJSONHandler(svc))
	r.Post("/updates/", handlers.UpdatesJSONHandler(svc, pub))

	return httptest.NewServer(r)
}

// ExampleMetricsHandler демонстрирует обновление gauge-метрики через сегменты
// URL (POST /update/{type}/{name}/{value}) и её чтение в текстовом виде
// (GET /value/{type}/{name}).
func ExampleMetricsHandler() {
	ts := newTestServer()
	defer ts.Close()

	// Обновляем метрику Temperature значением 36.6.
	resp, _ := http.Post(ts.URL+"/update/gauge/Temperature/36.6", "text/plain", nil)
	resp.Body.Close()
	fmt.Println("update status:", resp.StatusCode)

	// Читаем сохранённое значение обратно.
	resp, _ = http.Get(ts.URL + "/value/gauge/Temperature")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("value:", string(body))

	// Output:
	// update status: 200
	// value: 36.6
}

// ExampleUpdateJSONHandler демонстрирует обновление одной метрики в формате
// JSON (POST /update/). В ответе сервер возвращает применённую метрику.
func ExampleUpdateJSONHandler() {
	ts := newTestServer()
	defer ts.Close()

	value := 42.5
	body, _ := json.Marshal(models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &value})

	resp, _ := http.Post(ts.URL+"/update/", "application/json", bytes.NewReader(body))
	out, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Println(strings.TrimSpace(string(out)))

	// Output:
	// {"id":"Alloc","type":"gauge","value":42.5}
}

// ExampleUpdatesJSONHandler демонстрирует пакетное обновление метрик
// (POST /updates/) и последующее чтение counter-метрики в формате JSON
// (POST /value/).
func ExampleUpdatesJSONHandler() {
	ts := newTestServer()
	defer ts.Close()

	gauge := 12.0
	delta := int64(5)
	batch := []models.Metrics{
		{ID: "HeapAlloc", MType: models.Gauge, Value: &gauge},
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
	}
	body, _ := json.Marshal(batch)

	resp, _ := http.Post(ts.URL+"/updates/", "application/json", bytes.NewReader(body))
	fmt.Println("batch status:", resp.StatusCode)
	resp.Body.Close()

	// Запрашиваем значение счётчика PollCount обратно.
	query, _ := json.Marshal(models.Metrics{ID: "PollCount", MType: models.Counter})
	resp, _ = http.Post(ts.URL+"/value/", "application/json", bytes.NewReader(query))
	out, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Println(strings.TrimSpace(string(out)))

	// Output:
	// batch status: 200
	// {"id":"PollCount","type":"counter","delta":5}
}

// ExampleValueJSONHandler демонстрирует чтение ранее сохранённой gauge-метрики
// в формате JSON (POST /value/).
func ExampleValueJSONHandler() {
	ts := newTestServer()
	defer ts.Close()

	// Сначала сохраняем метрику.
	value := 3.14
	body, _ := json.Marshal(models.Metrics{ID: "Pi", MType: models.Gauge, Value: &value})
	resp, _ := http.Post(ts.URL+"/update/", "application/json", bytes.NewReader(body))
	resp.Body.Close()

	// Затем запрашиваем её значение.
	query, _ := json.Marshal(models.Metrics{ID: "Pi", MType: models.Gauge})
	resp, _ = http.Post(ts.URL+"/value/", "application/json", bytes.NewReader(query))
	out, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Println(strings.TrimSpace(string(out)))

	// Output:
	// {"id":"Pi","type":"gauge","value":3.14}
}
