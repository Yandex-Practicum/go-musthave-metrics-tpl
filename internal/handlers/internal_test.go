package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClientIP покрывает определение IP по заголовкам прокси и по RemoteAddr.
func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		realIP     string
		forwarded  string
		remoteAddr string
		want       string
	}{
		{name: "X-Real-IP", realIP: "10.0.0.1", remoteAddr: "1.2.3.4:5678", want: "10.0.0.1"},
		{name: "X-Forwarded-For", forwarded: "10.0.0.2, 10.0.0.3", remoteAddr: "1.2.3.4:5678", want: "10.0.0.2"},
		{name: "RemoteAddr", remoteAddr: "1.2.3.4:5678", want: "1.2.3.4"},
		{name: "RemoteAddr без порта", remoteAddr: "1.2.3.4", want: "1.2.3.4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.realIP != "" {
				r.Header.Set("X-Real-IP", tt.realIP)
			}
			if tt.forwarded != "" {
				r.Header.Set("X-Forwarded-For", tt.forwarded)
			}
			if got := clientIP(r); got != tt.want {
				t.Fatalf("clientIP() = %q, ожидалось %q", got, tt.want)
			}
		})
	}
}

// TestPingHandler_NoDB проверяет, что без соединения с БД возвращается 500.
func TestPingHandler_NoDB(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	PingHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("ожидался 500, получен %d", rec.Code)
	}
}
