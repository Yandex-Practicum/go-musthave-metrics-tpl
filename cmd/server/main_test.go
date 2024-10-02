package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestUpdateHandlerGauge(t *testing.T) {
    storage := &MemStorage{
        gauges:   make(map[string]float64),
        counters: make(map[string]int64),
    }

    req := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/42.5", nil)
    req.Header.Set("Content-Type", "text/plain")

    rr := httptest.NewRecorder()

    handler := updateHandler(storage)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")

    value, exists := storage.gauges["testGauge"]
    assert.True(t, exists, "Gauge testGauge should exist")
    assert.Equal(t, 42.5, value, "Gauge testGauge should have the value 42.5")
}

func TestUpdateHandlerCounter(t *testing.T) {
    storage := &MemStorage{
        gauges:   make(map[string]float64),
        counters: make(map[string]int64),
    }

    req := httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/10", nil)
    req.Header.Set("Content-Type", "text/plain")

    rr := httptest.NewRecorder()

    handler := updateHandler(storage)
    handler.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")

    value, exists := storage.counters["testCounter"]
    assert.True(t, exists, "Counter testCounter should exist")
    assert.Equal(t, int64(10), value, "Counter testCounter should have the value 10")

    req = httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/5", nil)
    rr = httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    value, exists = storage.counters["testCounter"]
    assert.True(t, exists, "Counter testCounter should exist")
    assert.Equal(t, int64(15), value, "Counter testCounter should have the value 15")
}
