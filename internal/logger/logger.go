package logger

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	loggerInstance *zap.Logger
	once           sync.Once
)

func InitLogger() *zap.Logger {
	once.Do(func() {
		config := zap.Config{
			Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
			OutputPaths: []string{"stdout", "app.log"},
			Encoding:    "json",
		}
		var err error
		loggerInstance, err = config.Build()
		if err != nil {
			log.Fatal("failed to initialize zap logger: " + err.Error())
		}
	})
	return loggerInstance
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := InitLogger() // Получаем логгер из синглтона
		start := time.Now()

		var requestBody []byte
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error("Error reading request body", zap.Error(err))
				http.Error(w, "Unable to read request body", http.StatusInternalServerError)
				return
			}
			requestBody = bodyBytes
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)
		contentType := r.Header.Get("Content-Type")

		logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("Content-Type", contentType),
			zap.String("body", string(requestBody)),
			zap.Int("status", lw.status),
			zap.Int("size", lw.size),
			zap.Duration("duration", duration),
		)
	})
}
