package logger_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"evgen3000/go-musthave-metrics-tpl.git/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	assert.NotPanics(t, func() {
		logger.InitLogger()
	}, "InitLogger should not panic")

	assert.NotNil(t, logger.GetLogger(), "Logger instance should not be nil after InitLogger is called")
}

func TestLoggingMiddleware(t *testing.T) {
	logger.InitLogger()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	wrappedHandler := logger.LoggingMiddleware(handler)
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("request body"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code to be 200")
	assert.Equal(t, "OK", string(body), "Expected response body to be 'OK'")
}

func TestLoggingMiddleware_RequestBodyLogging(t *testing.T) {
	logger.InitLogger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := r.Body.Close()
			if err != nil {
				panic(err)
			}
		}()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	wrappedHandler := logger.LoggingMiddleware(handler)
	reqBody := "sample request body"
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code to be 200")
}
