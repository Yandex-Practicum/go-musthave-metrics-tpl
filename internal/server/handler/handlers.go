package handler

import (
	"bytes"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vova4o/yandexadv/internal/models"
)

// StatisticPage обработчик для страницы статистики
func (s *Router) StatisticPage(c *gin.Context) {
    tmpl, metrics, err := s.Service.MetrixStatistic()
    if err != nil {
        c.String(http.StatusInternalServerError, "internal server error")
        return
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, metrics); err != nil {
        c.String(http.StatusInternalServerError, "internal server error")
        return
    }

    c.Header("Content-Type", "text/html")
    c.String(http.StatusOK, buf.String())
}

// UpdateMetricHandler обработчик для обновления метрики
func (s *Router) UpdateMetricHandler(c *gin.Context) {
    metric := models.Metric{
        Type:  c.Param("type"),
        Name:  c.Param("name"),
        Value: c.Param("value"),
    }
    
    err := s.Service.UpdateServ(metric)
    if err != nil {
        if httpErr, ok := err.(*models.HTTPError); ok {
            log.Printf("Error: %v", httpErr.Message)
            c.String(httpErr.Status, httpErr.Message)
            return
        }
        log.Printf("Internal server error: %v", err)
        c.String(http.StatusInternalServerError, "internal server error")
        return
    }

    c.Status(http.StatusOK)
}

// GetValueHandler обработчик для получения значения метрики
func (s *Router) GetValueHandler(c *gin.Context) {    
    metric := models.Metric{
        Type: c.Param("type"),
        Name: c.Param("name"),
    }
    
    value, err := s.Service.GetValueServ(metric)
    if err != nil {
        c.String(http.StatusNotFound, "metric not found")
        return
    }

    c.String(http.StatusOK, value)
}