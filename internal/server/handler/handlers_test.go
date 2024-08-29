package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vova4o/yandexadv/internal/models"
)

func TestUpdateMetricHandlerJSON(t *testing.T) {
    tests := []struct {
        name           string
        method         string
        url            string
        body           string
        updateFunc     func(metric models.Metrics) error
        wantStatusCode int
        wantBody       string
    }{
        {
            name:   "Valid update",
            method: http.MethodPost,
            url:    "/update",
            body:   `{"id":"requests","mtype":"counter","delta":10}`,
            updateFunc: func(metric models.Metrics) error {
                return nil
            },
            wantStatusCode: http.StatusOK,
            wantBody:       "",
        },
        // {
        //     name:   "Internal server error",
        //     method: http.MethodPost,
        //     url:    "/update",
        //     body:   `{"id":"requests","mtype":"counter","delta":10}`,
        //     updateFunc: func(metric models.Metrics) error {
        //         return errors.New("internal error")
        //     },
        //     wantStatusCode: http.StatusInternalServerError,
        //     wantBody:       "internal server error",
        // },
        // {
        //     name:           "Invalid request format",
        //     method:         http.MethodPost,
        //     url:            "/update",
        //     body:           `{"id":"requests","mtype":"counter"}`,
        //     updateFunc:     nil,
        //     wantStatusCode: http.StatusBadRequest,
        //     wantBody:       "bad request",
        // },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockService := &mockService{
                updateFuncJSON: tt.updateFunc,
            }

            r := gin.Default()
            router := New(mockService, nil)
            r.POST("/update", router.UpdateMetricHandlerJSON)

            req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
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

// func TestGetValueHandlerJSON(t *testing.T) {
//     tests := []struct {
//         name           string
//         body           string
//         expectedStatus int
//         expectedBody   string
//         mockValue      *models.Metrics
//         mockError      error
//     }{
//         {
//             name:           "Valid gauge metric",
//             body:           `{"id":"testGauge","mtype":"gauge"}`,
//             expectedStatus: http.StatusOK,
//             expectedBody:   `{"id":"testGauge","mtype":"gauge","value":123.45}`,
//             mockValue: &models.Metrics{
//                 ID:    "testGauge",
//                 MType: "gauge",
//                 Value: func() *float64 { v := 123.45; return &v }(),
//             },
//             mockError: nil,
//         },
//         {
//             name:           "Valid counter metric",
//             body:           `{"id":"testCounter","mtype":"counter"}`,
//             expectedStatus: http.StatusOK,
//             expectedBody:   `{"id":"testCounter","mtype":"counter","delta":678}`,
//             mockValue: &models.Metrics{
//                 ID:    "testCounter",
//                 MType: "counter",
//                 Delta: func() *int64 { v := int64(678); return &v }(),
//             },
//             mockError: nil,
//         },
//         // {
//         //     name:           "Metric not found",
//         //     body:           `{"id":"nonExistentMetric","mtype":"gauge"}`,
//         //     expectedStatus: http.StatusNotFound,
//         //     expectedBody:   `{"error":"metric not found"}`,
//         //     mockValue:      nil,
//         //     mockError:      errors.New("metric not found"),
//         // },
//     }

//     for _, tt := range tests {
//         t.Run(tt.name, func(t *testing.T) {
//             mockService := &mockService{
//                 getValueFuncJSON: func(metric models.Metrics) (*models.Metrics, error) {
//                     if metric.ID == tt.mockValue.ID && metric.MType == tt.mockValue.MType {
//                         return tt.mockValue, tt.mockError
//                     }
//                     return nil, tt.mockError
//                 },
//             }

//             r := gin.Default()
//             router := New(mockService, nil)
//             r.POST("/value", router.GetValueHandlerJSON)

//             req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.body))
//             req.Header.Set("Content-Type", "application/json")
//             w := httptest.NewRecorder()

//             r.ServeHTTP(w, req)

//             resp := w.Result()
//             defer resp.Body.Close()
//             body, _ := io.ReadAll(resp.Body)

//             assert.Equal(t, tt.expectedStatus, resp.StatusCode)
//             assert.JSONEq(t, tt.expectedBody, strings.TrimSpace(string(body)))
//         })
//     }
// }