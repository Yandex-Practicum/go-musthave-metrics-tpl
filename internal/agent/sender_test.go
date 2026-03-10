package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSender_SendGauge(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      float64
		serverCode int
		wantErr    bool
		wantPath   string
	}{
		{
			name:       "test #1 /200",
			metricName: "Alloc",
			value:      123,
			serverCode: http.StatusOK,
			wantErr:    false,
			wantPath:   "/update/gauge/Alloc/123",
		},
		{
			name:       "test #2 /500",
			metricName: "Alloc",
			value:      1,
			serverCode: http.StatusInternalServerError,
			wantErr:    true,
			wantPath:   "/update/gauge/Alloc/1",
		},
		{
			name:       "test #3 ",
			metricName: "Alloc",
			value:      -3.14,
			serverCode: http.StatusOK,
			wantErr:    false,
			wantPath:   "/update/gauge/Alloc/-3.14",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var Path, Method, ContentType string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Path = r.URL.RequestURI()
				Method = r.Method
				ContentType = r.Header.Get("Content-Type")
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
					t.Errorf("Ошибка: %v", err)
				}
			}
			if Path != tt.wantPath {
				t.Errorf("путь: ожидали %q, получили %q", tt.wantPath, Path)
			}
			if Method != http.MethodPost {
				t.Errorf("метод: ожидали POST, получили %q", Method)
			}
			if ContentType != "text/plain" {
				t.Errorf("Content-Type: ожидали text/plain, получили %q", ContentType)
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
		wantPath   string
	}{
		{
			name:       "test #1 /200",
			metricName: "PollCount",
			delta:      5,
			serverCode: http.StatusOK,
			wantErr:    false,
			wantPath:   "/update/counter/PollCount/5",
		},
		{
			name:       "test #2 /500",
			metricName: "PollCount",
			delta:      1,
			serverCode: http.StatusInternalServerError,
			wantErr:    true,
			wantPath:   "/update/counter/PollCount/1",
		},
		{
			name:       "test #3 ",
			metricName: "PollCount",
			delta:      -3,
			serverCode: http.StatusOK,
			wantErr:    false,
			wantPath:   "/update/counter/PollCount/-3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path, method, contentType string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path = r.URL.RequestURI()
				method = r.Method
				contentType = r.Header.Get("Content-Type")
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
					t.Errorf("неожиданная ошибка: %v", err)
				}
			}
			if path != tt.wantPath {
				t.Errorf("путь: ожидали %q, получили %q", tt.wantPath, path)
			}
			if method != http.MethodPost {
				t.Errorf("метод: ожидали POST, получили %q", method)
			}
			if contentType != "text/plain" {
				t.Errorf("Content-Type: ожидали text/plain, получили %q", contentType)
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
		wantPaths      []string
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
			wantPaths: []string{
				"/update/gauge/Alloc/100",
				"/update/gauge/Sys/200",
				"/update/counter/PollCount/3",
			},
		},
		{
			name: "test #2 ошибка сервера",
			gauges: []GaugeMetric{
				{Name: "Alloc", Value: 100},
			},
			pollCountDelta: 1,
			serverCode:     http.StatusInternalServerError,
			wantErr:        true,
			wantPaths:      []string{"/update/gauge/Alloc/100"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paths []string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.RequestURI())
				w.WriteHeader(tt.serverCode)
			}))
			defer srv.Close()

			sender := NewSender(srv.URL)
			err := sender.SendAll(tt.gauges, tt.pollCountDelta)

			if tt.wantErr {
				if err == nil {
					t.Error("успешно")
				}
			} else {
				if err != nil {
					t.Errorf("ошибка: %v", err)
				}
			}
			if len(paths) != len(tt.wantPaths) {
				t.Errorf("количество запросов: ожидали %d, получили %d", len(tt.wantPaths), len(paths))
				return
			}
			for i, wantPath := range tt.wantPaths {
				if paths[i] != wantPath {
					t.Errorf("запрос %d: ожидали путь %q, получили %q", i+1, wantPath, paths[i])
				}
			}
		})
	}
}
