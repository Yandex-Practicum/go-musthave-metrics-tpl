package handler

import (
	"bytes"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vova4o/yandexadv/internal/models"
)

// PingHandler обработчик для проверки подключения к базе данных
func (s *Router) PingHandler(c *gin.Context) {
	err := s.Service.PingDB()
	if err != nil {
		log.Printf("Failed to ping database: %v", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	c.String(http.StatusOK, "pong")
}

// GetValueHandlerJSON обработчик для передачи значения метрики в формате JSON
func (s *Router) GetValueHandlerJSON(c *gin.Context) {
	var metricReq models.Metrics

	// Парсинг JSON-запроса
	if err := c.ShouldBindJSON(&metricReq); err != nil {
		// log.Printf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// log.Printf("Received GET JSON request for metric: %v", metricReq)

	// Получение значения метрики
	metricResp, err := s.Service.GetValueServJSON(metricReq)
	if err != nil {
		if err == models.ErrMetricNotFound {
			// log.Printf("Metric not found: %v", err)
			c.String(http.StatusNotFound, "metric not found")
			return
		}
		// log.Printf("Failed to get updated value: %v", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	// log.Printf("Retrieved metric response: %v", metricResp)

	// Возвращение JSON-ответа с заполненными значениями метрик
	c.JSON(http.StatusOK, metricResp)
}

// UpdateMetricHandlerJSON обработчик для обновления метрики в формате JSON
func (s *Router) UpdateMetricHandlerJSON(c *gin.Context) {
	var metric models.Metrics
	if err := c.BindJSON(&metric); err != nil {
		// log.Printf("Failed to bind JSON: %v", err)
		c.String(http.StatusBadRequest, "bad request")
		return
	}

	// log.Printf("Received POST JSON metric for update: ID=%s, Type=%s, Delta=%v, Value=%v", metric.ID, metric.MType, metric.Delta, metric.Value)

	// // Преобразование указателей в значения
	// if metric.MType == "gauge" && metric.Value != nil {
	//     value := *metric.Value
	//     metric.Value = &value
	// } else if metric.MType == "counter" && metric.Delta != nil {
	//     delta := *metric.Delta
	//     metric.Delta = &delta
	// }

	err := s.Service.UpdateServJSON(&metric)

	if err != nil {
		if httpErr, ok := err.(*models.HTTPError); ok {
			// log.Printf("Error: %v", httpErr.Message)
			c.String(httpErr.Status, httpErr.Message)
			return
		}
		// log.Printf("Internal server error: %v", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	updatedVal, err := s.Service.GetValueServJSON(metric)
	if err != nil {
		if err == models.ErrMetricNotFound {
			// log.Printf("Metric not found: %v", err)
			c.String(http.StatusNotFound, "metric not found")
			return
		}
		// log.Printf("Failed to get updated value: %v", err)
		c.String(http.StatusInternalServerError, "internal server error")
		return
	}

	// log.Printf("Successfully updated metric: %v", updatedVal)

	c.JSON(http.StatusOK, updatedVal)
}

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
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	// log.Printf("Received POST TEXT update request for metric: type=%s, name=%s, value=%s", metricType, metricName, metricValue)

	var metric models.Metrics
	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			// log.Printf("Failed to parse gauge value: %v", err)
			c.String(http.StatusBadRequest, "invalid gauge value")
			return
		}
		metric = models.Metrics{
			ID:    metricName,
			MType: metricType,
			Value: &value,
		}
	case "counter":
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			// log.Printf("Failed to parse counter value: %v", err)
			c.String(http.StatusBadRequest, "invalid counter value")
			return
		}
		metric = models.Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &delta,
		}
	default:
		// log.Printf("Invalid metric type: %s", metricType)
		c.String(http.StatusBadRequest, "invalid metric type")
		return
	}

	err := s.Service.UpdateServJSON(&metric)
	if err != nil {
		// log.Printf("Failed to update metric: %v", err)
		c.String(http.StatusInternalServerError, "failed to update metric")
		return
	}

	// log.Printf("Successfully updated metric: %v", metric)
	c.Status(http.StatusOK)
}

// GetValueHandler обработчик для получения значения метрики
func (s *Router) GetValueHandler(c *gin.Context) {
	metric := models.Metrics{
		MType: c.Param("type"),
		ID:    c.Param("name"),
	}

	// log.Printf("Received GET TEXT request for metric: %v", metric)

	value, err := s.Service.GetValueServ(metric)
	if err != nil {
		// log.Printf("Failed to get value: %v", err)
		c.String(http.StatusNotFound, models.ErrMetricNotFound.Error())
		return
	}

	// log.Printf("Retrieved value for metric %s of type %s: %v", metric.ID, metric.MType, value)

	c.String(http.StatusOK, value)
}
