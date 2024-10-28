package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/handlers"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage/file_manager"
	"evgen3000/go-musthave-metrics-tpl.git/internal/dto"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func setupHandler() *handlers.Handler {
	fm := file_manager.FileManager{}
	memStorage := storage.NewMemStorage(storage.MemStorageConfig{
		StoreInterval:   5,
		FileStoragePath: "storage.json",
		Restore:         false,
	}, &fm)
	return handlers.NewHandler(memStorage)
}

func TestHomeHandler(t *testing.T) {
	h := setupHandler()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	h.HomeHandler(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "<h4>Gauges</h4>")
	assert.Contains(t, recorder.Body.String(), "<h4>Counters</h4>")
}

func TestUpdateMetricHandlerJSON(t *testing.T) {
	h := setupHandler()

	// Test for Counter metric
	counterMetric := dto.MetricsDTO{
		ID:    "test_counter",
		MType: handlers.MetricTypeCounter,
		Delta: new(int64),
	}
	*counterMetric.Delta = 42

	body, err := json.Marshal(counterMetric)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	h.UpdateMetricHandlerJSON(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "test_counter")
	assert.Contains(t, recorder.Body.String(), "42")

	// Test for Gauge metric
	gaugeMetric := dto.MetricsDTO{
		ID:    "test_gauge",
		MType: handlers.MetricTypeGauge,
		Value: new(float64),
	}
	*gaugeMetric.Value = 3.14

	body, err = json.Marshal(gaugeMetric)
	assert.NoError(t, err)

	req, err = http.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
	assert.NoError(t, err)
	recorder = httptest.NewRecorder()
	h.UpdateMetricHandlerJSON(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "test_gauge")
	assert.Contains(t, recorder.Body.String(), "3.14")
}

func TestUpdateMetricHandlerText(t *testing.T) {
	h := setupHandler()

	// Test for Counter metric
	req, err := http.NewRequest(http.MethodPost, "/update/counter/test_counter/42", nil)
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.UpdateMetricHandlerText)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	// Test for Gauge metric
	req, err = http.NewRequest(http.MethodPost, "/update/gauge/test_gauge/3.14", nil)
	assert.NoError(t, err)

	recorder = httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGetMetricHandlerJSON(t *testing.T) {
	h := setupHandler()

	h.Storage.SetGauge("test_gauge", 3.14)

	metric := dto.MetricsDTO{
		ID:    "test_gauge",
		MType: handlers.MetricTypeGauge,
	}
	body, err := json.Marshal(metric)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(body))
	assert.NoError(t, err)
	recorder := httptest.NewRecorder()
	h.GetMetricHandlerJSON(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "test_gauge")
	assert.Contains(t, recorder.Body.String(), "3.14")
}

func TestGetMetricHandlerText(t *testing.T) {
	h := setupHandler()

	// Add a metric to the storage for testing
	h.Storage.SetGauge("test_gauge", 3.14)

	req, err := http.NewRequest(http.MethodGet, "/value/gauge/test_gauge", nil)
	assert.NoError(t, err)

	r := chi.NewRouter()
	r.Get("/value/{metricType}/{metricName}", h.GetMetricHandlerText)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "3.14")
}
