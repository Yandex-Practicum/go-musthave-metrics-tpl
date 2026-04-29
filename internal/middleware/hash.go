package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
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

			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			receivedHash := r.Header.Get("HashSHA256")
			if receivedHash != "" {
				// читаем тело
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "ошибка при чтение тело запроса", http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewReader(body))

				h := hmac.New(sha256.New, []byte(key))
				h.Write(body)
				expectedHash := hex.EncodeToString(h.Sum(nil))

				if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
					http.Error(w, "hash несоотвествует", http.StatusBadRequest)
					return
				}
			}

			hw := &hashResponseWriter{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(hw, r)

			h := hmac.New(sha256.New, []byte(key))
			h.Write(hw.body.Bytes())
			w.Header().Set("HashSHA256", hex.EncodeToString(h.Sum(nil)))
		})
	}
}
