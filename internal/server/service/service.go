package service

import (
	"errors"
	"fmt"
	"html/template"
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
	UpdateMetric(metric models.Metrics) error
	GetValue(metric models.Metrics) (*models.Metrics, error)
	MetrixStatistic() (map[string]models.Metrics, error)
}

// New создание нового сервиса
func New(s Storager) *Service {
	return &Service{
		Storage: s,
	}
}

// GetValueServJSON получение значения метрики в формате JSON
func (s *Service) GetValueServJSON(metric models.Metrics) (*models.Metrics, error) {
	// Проверка метрики
	if err := validateMetricJSON(&metric); err != nil {
		return nil, err
	}

	value, err := s.Storage.GetValue(metric)
	if err != nil {
		log.Printf("failed to get value: %v", err)
		return nil, err
	}

	return value, nil

}

// UpdateServJSON обновление метрики в формате JSON
func (s *Service) UpdateServJSON(metric *models.Metrics) error {
	// Проверка метрики
	if err := validateMetricJSON(metric); err != nil {
		return err
	}

	switch metric.MType {
	case "gauge":
		s.Storage.UpdateMetric(models.Metrics{
			MType: metric.MType,
			ID:    metric.ID,
			Value: metric.Value,
		})

	case "counter":
		// Получение старого значения счетчика
		counterVal, err := s.GetValueServ(models.Metrics{
			MType: metric.MType,
			ID:    metric.ID,
		})
		if err != nil {
			if errors.Is(err, models.ErrMetricNotFound) {
				counterVal = "0"
			} else {
				return err
			}
		}

		if counterVal == "" {
			counterVal = "0"
		}

		counterInt, err := strconv.Atoi(counterVal)
		if err != nil {
			log.Printf("failed to convert value to int: %v", err)
			return models.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to int: %v", err))
		}

		// Добавление старого значения к новому
		totalValue := *metric.Delta + int64(counterInt)
		s.Storage.UpdateMetric(models.Metrics{
			MType: metric.MType,
			ID:    metric.ID,
			Delta: &totalValue,
		})
	default:
		log.Printf("unknown metric type: %s", metric.MType)
		return models.NewHTTPError(http.StatusBadRequest, "unknown metric type")
	}

	return nil
}

// MetrixStatistic получение статистики метрик
func (s *Service) MetrixStatistic() (*template.Template, map[string]models.Metrics, error) {
	metrics, err := s.Storage.MetrixStatistic()
	if err != nil {
		log.Printf("failed to get metrics: %v", err)
		return nil, nil, models.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get metrics: %v", err))
	}

	tmpl, err := template.New("metrics").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Metrics Statistics</title>
		</head>
		<body>
			<h1>Metrics Statistics</h1>
			<table border="1">
				<tr>
					<th>Metric Name</th>
					<th>Metric Value</th>
				</tr>
				{{range $key, $metric := .}}
				<tr>
					<td>{{$key}}</td>
					<td>
						{{if eq $metric.MType "gauge"}}
							{{$metric.Value}}
						{{else}}
							{{$metric.Delta}}
						{{end}}
					</td>
				</tr>
				{{end}}
			</table>
		</body>
		</html>
	`)
	if err != nil {
		log.Printf("failed to parse template: %v", err)
		return nil, nil, models.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to parse template: %v", err))
	}

	return tmpl, metrics, nil
}

// GetValueServ получение значения метрики
func (s *Service) GetValueServ(metric models.Metrics) (string, error) {
	// Проверка метрики
	if err := validateMetricJSON(&metric); err != nil {
		return "", err
	}

	value, err := s.Storage.GetValue(metric)
	if err != nil {
		log.Printf("failed to get value: %v", err)
		return "", err
	}

	var valueStr string
	switch metric.MType {
	case "gauge":
		if value.Value != nil {
			valueStr = fmt.Sprintf("%v", *value.Value)
		}
	case "counter":
		if value.Delta != nil {
			valueStr = fmt.Sprintf("%v", *value.Delta)
		}
	default:
		return "", fmt.Errorf("unsupported metric type: %s", metric.MType)
	}

	return valueStr, nil
}

// UpdateServ обновление метрики
func (s *Service) UpdateServ(metric models.Metric) error {
	// Проверка метрики
	if err := validateMetric(metric); err != nil {
		return err
	}

	switch metric.Type {
	case "gauge":
		valueStr, ok := metric.Value.(string)
		if !ok {
			log.Printf("failed to assert value to string: %v", metric.Value)
			return models.NewHTTPError(http.StatusInternalServerError, "failed to assert value to string")
		}

		valueFloat, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			log.Printf("failed to convert value to float: %v", err)
			return models.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to float: %v", err))
		}

		s.Storage.UpdateMetric(models.Metrics{
			MType: metric.Type,
			ID:    metric.Name,
			Value: &valueFloat,
		})

	case "counter":
		// Обработка для типа counter
		valueStr, ok := metric.Value.(string)
		if !ok {
			log.Printf("failed to assert value to string: %v", metric.Value)
			return models.NewHTTPError(http.StatusInternalServerError, "failed to assert value to string")
		}

		valueInt, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			log.Printf("failed to convert value to int64: %v", err)
			return models.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to int64: %v", err))
		}

		// Получение старого значения счетчика
		counterVal, err := s.GetValueServ(models.Metrics{
			MType: metric.Type,
			ID:    metric.Name,
		})
		if err != nil {
			if errors.Is(err, models.ErrMetricNotFound) {
				counterVal = "0"
			} else {
				return err
			}
		}

		counterInt, err := strconv.ParseInt(counterVal, 10, 64)
		if err != nil {
			log.Printf("failed to convert value to int64: %v", err)
			return models.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to int64: %v", err))
		}

		// Добавление старого значения к новому
		totalValue := valueInt + counterInt
		s.Storage.UpdateMetric(models.Metrics{
			MType: metric.Type,
			ID:    metric.Name,
			Delta: &totalValue,
		})

	default:
		return models.NewHTTPError(http.StatusBadRequest, "unsupported metric type")
	}

	return nil
}

// validateMetric проверяет метрику на наличие ошибок
func validateMetric(metric models.Metric) error {
	if metric.Type == "" || metric.Value == "" || metric.Name == "" {
		log.Println("metric Values cannot be empty")
		return models.NewHTTPError(http.StatusBadRequest, "metricType cannot be empty")
	}

	return nil
}

// validateMetric проверяет метрику на наличие ошибок
func validateMetricJSON(metric *models.Metrics) error {
	if metric.MType == "" {
		log.Println("metricType cannot be empty")
		return models.NewHTTPError(http.StatusBadRequest, "metricType cannot be empty")
	}

	if metric.ID == "" {
		log.Println("metricIDcannot be empty")
		return models.NewHTTPError(http.StatusNotFound, "metricName cannot be empty")
	}

	return nil
}
