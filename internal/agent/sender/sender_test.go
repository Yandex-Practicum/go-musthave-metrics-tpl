package sender_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vova4o/yandexadv/internal/agent/flags"
	"github.com/vova4o/yandexadv/internal/agent/metrics"
	"github.com/vova4o/yandexadv/internal/agent/sender"
)

func TestCompressData(t *testing.T) {
	data := []byte("test data")
	compressedData, err := sender.CompressData(data)
	assert.NoError(t, err)

	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	assert.NoError(t, err)
	defer reader.Close()

	decompressedData, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, data, decompressedData)
}

func TestServerSupportsGzip(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Encoding") == "gzip" {
			w.Header().Set("Content-Encoding", "gzip")
		}
		w.WriteHeader(http.StatusOK)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	supportsGzip := sender.ServerSupportsGzip(server.Listener.Addr().String())
	assert.True(t, supportsGzip)
}

var config = &flags.Config{
	ServerAddress: "test_server",
	SecretKey:     "test_key",
}

func TestSendMetricsBatch(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var metricsData []metrics.Metrics
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			assert.NoError(t, err)
			defer reader.Close()
			err = json.NewDecoder(reader).Decode(&metricsData)
			assert.NoError(t, err)
		} else {
			err := json.NewDecoder(r.Body).Decode(&metricsData)
			assert.NoError(t, err)
		}
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	metricsData := []metrics.Metrics{
		{ID: "metric1", Value: float64Ptr(10)},
		{ID: "metric2", Value: float64Ptr(20)},
	}

	sender.SendMetricsBatch(config, metricsData)
}

func TestSendMetrics(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	metricsData := []metrics.Metrics{
		{ID: "metric1", Value: float64Ptr(10)},
		{ID: "metric2", Delta: int64Ptr(20)},
	}

	sender.SendMetrics(server.URL, metricsData)
}

func TestSendMetricsJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var metric metrics.Metrics
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			assert.NoError(t, err)
			defer reader.Close()
			err = json.NewDecoder(reader).Decode(&metric)
			assert.NoError(t, err)
		} else {
			err := json.NewDecoder(r.Body).Decode(&metric)
			assert.NoError(t, err)
		}
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	metricsData := []metrics.Metrics{
		{ID: "metric1", Value: float64Ptr(10)},
		{ID: "metric2", Delta: int64Ptr(20)},
	}

	sender.SendMetricsJSON(server.URL, metricsData)
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
