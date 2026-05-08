package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/middleware/hash"
)

type hashResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func HashCheckMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHash := r.Header.Get("HashSHA256")
			if receivedHash != "" {
				// читаем тело
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "ошибка при чтение тело запроса", http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(body))

				if !hash.Verify(body, key, receivedHash) {
					http.Error(w, "hash не соответствует", http.StatusBadRequest)
					return
				}

				hw := &hashResponseWriter{
					ResponseWriter: w,
					body:           &bytes.Buffer{},
				}

				next.ServeHTTP(hw, r)

				w.Header().Set("HashSHA256", hash.ComputeHMAC(hw.body.Bytes(), key))
			}
		})
	}
}
