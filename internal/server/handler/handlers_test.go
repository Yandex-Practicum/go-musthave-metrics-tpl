package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vova4o/yandexadv/internal/models"
)

func TestRouter_UpdateMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		updateFunc     func(metric models.Metric) error
		wantStatusCode int
		wantBody       string
	}{
		{
			name:   "Valid update",
			method: http.MethodPost,
			url:    "/update/counter/requests/10",
			updateFunc: func(metric models.Metric) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "",
		},
		{
			name:   "Internal server error",
			method: http.MethodPost,
			url:    "/update/counter/requests/10",
			updateFunc: func(metric models.Metric) error {
				return errors.New("internal error")
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       "internal server error",
		},
		// {
		//     name:           "Invalid request method",
		//     method:         http.MethodGet,
		//     url:            "/update/counter/requests/10",
		//     updateFunc:     nil,
		//     wantStatusCode: http.StatusMethodNotAllowed,
		//     wantBody:       "Invalid request method",
		// },
		// {
		//     name:           "Invalid request format",
		//     method:         http.MethodPost,
		//     url:            "/update/counter/requests",
		//     updateFunc:     nil,
		//     wantStatusCode: http.StatusNotFound,
		//     wantBody:       "Invalid request format",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockService{
				updateFunc: tt.updateFunc,
			}

			r := gin.Default()
			router := New(mockService)
			r.Any("/update/*path", router.UpdateMetricHandler)

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(string(body)))
		})
	}
}

func TestGetValueHandler(t *testing.T) {
	// Создаем тестовый сервис
	mockService := &mockService{}

	// Определяем тестовые случаи
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		expectedStatus int
		expectedBody   string
		mockValue      string
		mockError      error
	}{
		{
			name:           "Valid gauge metric",
			metricType:     "gauge",
			metricName:     "testGauge",
			expectedStatus: http.StatusOK,
			expectedBody:   "123.45",
			mockValue:      "123.45",
			mockError:      nil,
		},
		{
			name:           "Valid counter metric",
			metricType:     "counter",
			metricName:     "testCounter",
			expectedStatus: http.StatusOK,
			expectedBody:   "678",
			mockValue:      "678",
			mockError:      nil,
		},
		{
			name:           "Metric not found",
			metricType:     "gauge",
			metricName:     "nonExistentMetric",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "metric not found",
			mockValue:      "",
			mockError:      errors.New("metric not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем mockService
			mockService.MocGetValueServ = func(metric models.Metric) (string, error) {
				if metric.Type == tt.metricType && metric.Name == tt.metricName {
					return tt.mockValue, tt.mockError
				}
				return "", nil
			}

			// Создаем маршрутизатор
			r := gin.Default()
			router := New(mockService)
			r.GET("/value/:type/:name", router.GetValueHandler)

			// Создаем HTTP-запрос
			req := httptest.NewRequest("GET", "/value/"+tt.metricType+"/"+tt.metricName, nil)
			w := httptest.NewRecorder()

			// Выполняем запрос
			r.ServeHTTP(w, req)

			// Проверяем статус код и тело ответа
			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, tt.expectedBody, strings.TrimSpace(string(body)))
		})
	}
}
