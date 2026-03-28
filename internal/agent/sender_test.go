package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
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
				json.NewDecoder(r.Body).Decode(&received)
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
				json.NewDecoder(r.Body).Decode(&received)
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
			wantCount:      1,
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
