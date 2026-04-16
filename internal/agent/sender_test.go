package agent

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestSender_SendGauge(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      float64
		serverCode int
		wantErr    bool
	}{
		{
			name:       "test #1 /200",
			metricName: "Alloc",
			value:      123,
			serverCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "test #2 /500",
			metricName: "Alloc",
			value:      1,
			serverCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "test #3",
			metricName: "Alloc",
			value:      -3.14,
			serverCode: http.StatusOK,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var method, contentType, path string
			var received models.Metrics

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path = r.URL.RequestURI()
				method = r.Method
				contentType = r.Header.Get("Content-Type")
				gr, _ := gzip.NewReader(r.Body)
				if gr != nil {
					defer gr.Close()
					json.NewDecoder(gr).Decode(&received)
				}
				w.WriteHeader(tt.serverCode)
			}))
			defer srv.Close()

			sender := NewSender(srv.URL)
			err := sender.SendGauge(tt.metricName, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("ожидали ошибку, но её нет")
				}
			} else {
				if err != nil {
					t.Errorf("ошибка: %v", err)
				}
			}
			if path != "/update/" {
				t.Errorf("путь: ожидали %q, получили %q", "/update/", path)
			}
			if method != http.MethodPost {
				t.Errorf("метод: ожидали POST, получили %q", method)
			}
			if contentType != "application/json" {
				t.Errorf("Content-Type: ожидали application/json, получили %q", contentType)
			}
			if received.ID != tt.metricName {
				t.Errorf("ID: ожидали %q, получили %q", tt.metricName, received.ID)
			}
			if received.MType != "gauge" {
				t.Errorf("MType: ожидали %q, получили %q", "gauge", received.MType)
			}
			if received.Value != nil && *received.Value != tt.value {
				t.Errorf("Value: ожидали %v, получили %v", tt.value, *received.Value)
			}
		})
	}
}

func TestSender_SendCounter(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		delta      int64
		serverCode int
		wantErr    bool
	}{
		{
			name:       "test #1 /200",
			metricName: "PollCount",
			delta:      5,
			serverCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "test #2 /500",
			metricName: "PollCount",
			delta:      1,
			serverCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "test #3",
			metricName: "PollCount",
			delta:      -3,
			serverCode: http.StatusOK,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path, method, contentType string
			var received models.Metrics

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path = r.URL.RequestURI()
				method = r.Method
				contentType = r.Header.Get("Content-Type")
				gr, _ := gzip.NewReader(r.Body)
				if gr != nil {
					defer gr.Close()
					json.NewDecoder(gr).Decode(&received)
				}
				w.WriteHeader(tt.serverCode)
			}))
			defer srv.Close()

			sender := NewSender(srv.URL)
			err := sender.SendCounter(tt.metricName, tt.delta)

			if tt.wantErr {
				if err == nil {
					t.Error("ожидали ошибку, но её нет")
				}
			} else {
				if err != nil {
					t.Errorf("ошибка: %v", err)
				}
			}
			if path != "/update/" {
				t.Errorf("путь: ожидали %q, получили %q", "/update/", path)
			}
			if method != http.MethodPost {
				t.Errorf("метод: ожидали POST, получили %q", method)
			}
			if contentType != "application/json" {
				t.Errorf("Content-Type: ожидали application/json, получили %q", contentType)
			}
			if received.ID != tt.metricName {
				t.Errorf("ID: ожидали %q, получили %q", tt.metricName, received.ID)
			}
			if received.MType != "counter" {
				t.Errorf("MType: ожидали %q, получили %q", "counter", received.MType)
			}
			if received.Delta != nil && *received.Delta != tt.delta {
				t.Errorf("Delta: ожидали %v, получили %v", tt.delta, *received.Delta)
			}
		})
	}
}

func TestSender_SendAll(t *testing.T) {
	tests := []struct {
		name           string
		gauges         []GaugeMetric
		pollCountDelta int64
		serverCode     int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "test #1 успех",
			gauges: []GaugeMetric{
				{Name: "Alloc", Value: 100},
				{Name: "Sys", Value: 200},
			},
			pollCountDelta: 3,
			serverCode:     http.StatusOK,
			wantErr:        false,
			wantCount:      3,
		},
		{
			name: "test #2 ошибка сервера",
			gauges: []GaugeMetric{
				{Name: "Alloc", Value: 100},
			},
			pollCountDelta: 1,
			serverCode:     http.StatusInternalServerError,
			wantErr:        true,
			wantCount:      4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var count int

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				w.WriteHeader(tt.serverCode)
			}))
			defer srv.Close()

			sender := NewSender(srv.URL)
			err := sender.SendAll(tt.gauges, tt.pollCountDelta)

			if tt.wantErr {
				if err == nil {
					t.Error("ожидали ошибку")
				}
			} else {
				if err != nil {
					t.Errorf("ошибка: %v", err)
				}
			}
			if count != tt.wantCount {
				t.Errorf("количество запросов: ожидали %d, получили %d", tt.wantCount, count)
			}
		})
	}
}

func TestSender_SendBatch(t *testing.T) {
	tests := []struct {
		name       string
		gauges     []GaugeMetric
		delta      int64
		serverCode int
		wantErr    bool
		wantCalled bool
		wantCount  int // ожидаемое кол-во метрик в теле
	}{
		{
			name:       "Успех",
			gauges:     []GaugeMetric{{Name: "Alloc", Value: 100}, {Name: "Sys", Value: 200}},
			delta:      5,
			serverCode: http.StatusOK,
			wantErr:    false,
			wantCalled: true,
			wantCount:  3, // 2 gauge + 1 counter
		},
		{
			name:       "Пустой",
			gauges:     []GaugeMetric{},
			delta:      0,
			serverCode: http.StatusOK,
			wantErr:    false,
			wantCalled: false, // HTTP не должен вызываться
			wantCount:  0,
		},
		{
			name:       "Ошибка сервера",
			gauges:     []GaugeMetric{{Name: "X", Value: 1}},
			delta:      1,
			serverCode: http.StatusInternalServerError,
			wantErr:    true,
			wantCalled: true,
			wantCount:  2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var contentEncoding, contentType string
			var metrics []models.Metrics

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				contentEncoding = r.Header.Get("Content-Encoding")
				contentType = r.Header.Get("Content-Type")

				gr, err := gzip.NewReader(r.Body)
				if err == nil {
					defer gr.Close()
					body, _ := io.ReadAll(gr)
					json.Unmarshal(body, &metrics)
				}
				w.WriteHeader(tt.serverCode)
			}))
			defer srv.Close()

			sender := NewSender(srv.URL)
			err := sender.SendBatch(tt.gauges, tt.delta)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantCalled, called)

			if tt.wantCalled {
				assert.Equal(t, "gzip", contentEncoding)
				assert.Equal(t, "application/json", contentType)
				assert.Len(t, metrics, tt.wantCount)
			}
		})
	}
}
