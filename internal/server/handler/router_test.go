package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vova4o/yandexadv/internal/models"
)

// Предположим, что интерфейс Service определен так:
type Storager interface {
    Update(metricType, metricName string, metricValue interface{}) error
}

// mockService должен реализовывать этот интерфейс
type mockService struct {
    updateFunc func(metricType, metricName string, metricValue interface{}) error
}

func (m *mockService) Update(metricType, metricName string, metricValue interface{}) error {
    if m.updateFunc != nil {
        return m.updateFunc(metricType, metricName, metricValue)
    }
    return nil
}

func TestRouter_UpdateMetricHandler(t *testing.T) {
    tests := []struct {
        name           string
        method         string
        url            string
        updateFunc     func(metricType, metricName string, metricValue interface{}) error
        wantStatusCode int
        wantBody       string
    }{
        {
            name:           "Invalid request method",
            method:         http.MethodGet,
            url:            "/update/counter/requests/10",
            updateFunc:     nil,
            wantStatusCode: http.StatusMethodNotAllowed,
            wantBody:       "Invalid request method\n",
        },
        {
            name:           "Invalid request format",
            method:         http.MethodPost,
            url:            "/update/counter/requests",
            updateFunc:     nil,
            wantStatusCode: http.StatusNotFound,
            wantBody:       "Invalid request format\n",
        },
        {
            name:   "Service update error",
            method: http.MethodPost,
            url:    "/update/counter/requests/10",
            updateFunc: func(metricType, metricName string, metricValue interface{}) error {
                return &models.HTTPError{
                    Message: "Invalid metric value",
                    Status:  http.StatusBadRequest,
                }
            },
            wantStatusCode: http.StatusBadRequest,
            wantBody:       "Invalid metric value\n",
        },
        {
            name:   "Internal server error",
            method: http.MethodPost,
            url:    "/update/counter/requests/10",
            updateFunc: func(metricType, metricName string, metricValue interface{}) error {
                return &models.HTTPError{
                    Message: "Internal server error",
                    Status:  http.StatusInternalServerError,
                }
            },
            wantStatusCode: http.StatusInternalServerError,
            wantBody:       "Internal server error\n",
        },
        {
            name:   "Valid update",
            method: http.MethodPost,
            url:    "/update/counter/requests/10",
            updateFunc: func(metricType, metricName string, metricValue interface{}) error {
                return nil
            },
            wantStatusCode: http.StatusOK,
            wantBody:       "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := &mockService{
                updateFunc: tt.updateFunc,
            }

            router := New(mockService)
            router.RegisterRoutes()

            req := httptest.NewRequest(tt.method, tt.url, nil)
            w := httptest.NewRecorder()

            router.UpdateMetricHandler()(w, req)

            resp := w.Result()
            defer resp.Body.Close()
            body, _ := io.ReadAll(resp.Body)

            if resp.StatusCode != tt.wantStatusCode {
                t.Errorf("got status %d, want %d", resp.StatusCode, tt.wantStatusCode)
            }

            if string(body) != tt.wantBody {
                t.Errorf("got body %q, want %q", body, tt.wantBody)
            }
        })
    }
}