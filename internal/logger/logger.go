package logger

import (
	"bytes"
	"io"
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

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func InitLogger() {
	once.Do(func() {
		config := zap.Config{Level: zap.NewAtomicLevelAt(zapcore.InfoLevel), OutputPaths: []string{"stdout", "app.log"}, Encoding: "json"}
		var err error
		loggerInstance, err = config.Build()
		if err != nil {
			panic("failed to initialize zap logger: " + err.Error())
		}
	})

}

func GetLogger() *zap.Logger {
	if loggerInstance == nil {
		InitLogger()
	}
	return loggerInstance
}

func HandlerLog(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var requestBody []byte
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				GetLogger().Error("Error reading request body", zap.Error(err))
				http.Error(w, "Unable to read request body", http.StatusInternalServerError)
				return
			}
			requestBody = bodyBytes

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		contentType := r.Header.Get("Content-Type")

		GetLogger().Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.String("Content-Type", contentType),
			zap.String("body", string(requestBody)),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
			zap.Duration("duration", duration),
		)
	}
	return http.HandlerFunc(logFn)
}
