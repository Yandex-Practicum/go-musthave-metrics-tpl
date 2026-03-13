package logger

import (
	"net/http"
	"time"
)

type responseLogger struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rl *responseLogger) WriteHeader(code int) {
	rl.statusCode = code
	rl.ResponseWriter.WriteHeader(code)
}

func (rl *responseLogger) Write(b []byte) (int, error) {
	if rl.statusCode == 0 {
		rl.statusCode = http.StatusOK
	}
	n, err := rl.ResponseWriter.Write(b)
	rl.size += n
	return n, err
}

// Middleware логирует запросы и ответы (Info уровень)
func Middleware(next http.Handler) http.Handler {
	logger := GetLogger()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rl := &responseLogger{ResponseWriter: w}

		next.ServeHTTP(rl, r)

		duration := time.Since(start)

		logger.Info().
			Str("method", r.Method).
			Str("uri", r.RequestURI).
			Int("status", rl.statusCode).
			Int("size", rl.size).
			Dur("duration_ms", duration).
			Msg("HTTP request handled")
	})
}
