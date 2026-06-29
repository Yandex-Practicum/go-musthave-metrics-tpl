package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/middleware/hash"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status":"ok"}`)
	})
}

func TestGzipMiddleware_CompressesResponse(t *testing.T) {
	h := GzipMiddleware(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("ожидался Content-Encoding gzip, получено %q", rec.Header().Get("Content-Encoding"))
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("тело должно быть валидным gzip: %v", err)
	}
	data, _ := io.ReadAll(gr)
	if string(data) != `{"status":"ok"}` {
		t.Fatalf("неожиданное тело: %s", data)
	}
}

func TestGzipMiddleware_DecompressesRequest(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("hello"))
	gw.Close()

	var got string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()
	GzipMiddleware(next).ServeHTTP(rec, req)

	if got != "hello" {
		t.Fatalf("тело должно быть распаковано, получено %q", got)
	}
}

func TestGzipMiddleware_PassThrough(t *testing.T) {
	rec := httptest.NewRecorder()
	GzipMiddleware(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Fatal("без Accept-Encoding не должно быть gzip")
	}
}

func TestHashCheckMiddleware_ValidHash(t *testing.T) {
	key := "secret"
	body := []byte(`{"id":"x"}`)
	sum := hash.ComputeHMAC(body, key)

	h := HashCheckMiddleware(key)(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", sum)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ожидался 200, получено %d", rec.Code)
	}
	if rec.Header().Get("HashSHA256") == "" {
		t.Fatal("ответ должен содержать HashSHA256")
	}
}

func TestHashCheckMiddleware_InvalidHash(t *testing.T) {
	h := HashCheckMiddleware("secret")(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("data")))
	req.Header.Set("HashSHA256", "deadbeef")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("ожидался 400 на неверный hash, получено %d", rec.Code)
	}
}
