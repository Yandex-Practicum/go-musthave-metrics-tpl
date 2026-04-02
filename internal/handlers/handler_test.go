package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRouter(repo storage.Repository) http.Handler {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", MetricsHandler(repo))
	r.Get("/value/{type}/{name}", ValueHandler(repo))
	r.Get("/", ListHandler(repo))
	return r
}

func TestMetricsHandler(t *testing.T) {
	repo := storage.NewMemoryStorage()
	ts := httptest.NewServer(newTestRouter(repo))
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/update/gauge/cpu/3.14", "text/plain", nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Post(ts.URL+"/update/counter/hits/10", "text/plain", nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Post(ts.URL+"/update/unknown/cpu/1", "text/plain", nil)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestValueHandler(t *testing.T) {
	repo := storage.NewMemoryStorage()
	repo.UpdateGauge(context.Background(), "cpu", 3.14)
	repo.UpdateCounter(context.Background(), "hits", 10)
	ts := httptest.NewServer(newTestRouter(repo))
	defer ts.Close()

	// получить gauge
	resp, err := http.Get(ts.URL + "/value/gauge/cpu")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "3.14", string(body))

	// получить counter
	resp, err = http.Get(ts.URL + "/value/counter/hits")
	require.NoError(t, err)
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "10", string(body))

	// неизвестная метрика
	resp, err = http.Get(ts.URL + "/value/gauge/unknown")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListHandler(t *testing.T) {
	repo := storage.NewMemoryStorage()
	repo.UpdateGauge(context.Background(), "cpu", 1.5)
	ts := httptest.NewServer(newTestRouter(repo))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), "cpu")
}
