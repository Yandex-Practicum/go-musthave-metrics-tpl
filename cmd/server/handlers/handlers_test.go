package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/router"
	"evgen3000/go-musthave-metrics-tpl.git/cmd/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandlerGaugeSuccess(t *testing.T) {
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/temperature/23.5", nil)
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	value, exists := s.GetGauge("temperature")
	assert.True(t, exists)
	assert.Equal(t, 23.5, value)
}

func TestUpdateHandlerCounterSuccess(t *testing.T) {
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/hits/10", nil)
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	value, exists := s.GetCounter("hits")
	assert.True(t, exists)
	assert.Equal(t, int64(10), value)
}

func TestUpdateHandlerInvalidMetricType(t *testing.T) {
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)

	req := httptest.NewRequest(http.MethodPost, "/update/unknown/temperature/23.5", nil)
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetHandlerGaugeSuccess(t *testing.T) {
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)
	s.SetGauge("temperature", 23.5)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/temperature", nil)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "23.5", rr.Body.String())
}

func TestGetHandlerCounterSuccess(t *testing.T) {
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)
	s.IncrementCounter("hits", 10)

	req := httptest.NewRequest(http.MethodGet, "/value/counter/hits", nil)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "10", rr.Body.String())
}

func TestGetHandlerMetricNotFound(t *testing.T) {
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/unknown", nil)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHomeHandle(t *testing.T) {
	s := storage.NewMemStorage()
	r := router.SetupRouter(s)
	s.SetGauge("temperature", 23.5)
	s.IncrementCounter("hits", 10)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()

	assert.Contains(t, body, "Gauges")
	assert.Contains(t, body, "temperature: 23.5")
	assert.Contains(t, body, "Counters")
	assert.Contains(t, body, "hits: 10")
}
