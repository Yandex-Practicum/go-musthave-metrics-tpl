package service

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/vova4o/yandexadv/internal/models"
)

// Service структура для бизнес-логики
type Service struct {
	Storage Storager
}

// Storager интерфейс для хранилища
type Storager interface {
	Update(metricType, metricName string, metricValue interface{}) error
}

// New создание нового сервиса
func New(s Storager) *Service {
	return &Service{
		Storage: s,
	}
}

// Update обновление метрики
func (s *Service) Update(metricType, metricName string, metricValue interface{}) error {
	// TODO: make it smaller in one place
	if metricType == "" {
		log.Println("metricType cannot be empty")
		return models.NewHTTPError(http.StatusBadRequest, "metricType cannot be empty")
	}

	if metricName == "" {
		log.Println("metricName cannot be empty")
		return models.NewHTTPError(http.StatusNotFound, "metricName cannot be empty")
	}

	if metricValue == "" {
		log.Println("metricValue cannot be nil")
		return models.NewHTTPError(http.StatusBadRequest, "metricValue cannot be nil")
	}

	if s.Storage == nil {
		log.Println("storage cannot be nil")
		return models.NewHTTPError(http.StatusInternalServerError, "storage cannot be nil")
	}

	switch metricType {
    case "gauge":
        strValue, ok := metricValue.(string)
        if !ok {
            log.Println("metricValue must be a string for gauge type")
            return models.NewHTTPError(http.StatusBadRequest, "metricValue must be a string for gauge type")
        }
        value, err := strconv.ParseFloat(strValue, 64)
        if err != nil {
            log.Printf("invalid gauge value: %v", err)
            return models.NewHTTPError(http.StatusBadRequest, "invalid gauge value")
        }
        s.Storage.Update(metricType, metricName, value)
    case "counter":
        strValue, ok := metricValue.(string)
        if !ok {
            log.Println("metricValue must be a string for counter type")
            return models.NewHTTPError(http.StatusBadRequest, "metricValue must be a string for counter type")
        }
        value, err := strconv.ParseInt(strValue, 10, 64)
        if err != nil {
            log.Printf("invalid counter value: %v", err)
            return models.NewHTTPError(http.StatusBadRequest, "invalid counter value")
        }
        s.Storage.Update(metricType, metricName, value)
    default:
        log.Printf("unknown metric type: %s", metricType)
        return models.NewHTTPError(http.StatusBadRequest, "unknown metric type")
    }

	if err := s.Storage.Update(metricType, metricName, metricValue); err != nil {
		log.Printf("failed to update metric: %v", err)
		return models.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to update metric: %v", err))
	}

	return nil
}

