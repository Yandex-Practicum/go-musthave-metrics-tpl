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
	UpdateMetric(metric models.Metric) error
	GetValue(metric models.Metric) (interface{}, error)
	MetrixStatistic() (map[string]interface{}, error)
}

// New создание нового сервиса
func New(s Storager) *Service {
	return &Service{
		Storage: s,
	}
}

// MetrixStatistic получение статистики метрик
func (s *Service) MetrixStatistic() (*template.Template, map[string]interface{}, error) {
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
					<td>{{$metric.Value}}</td>
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

// 	// TODO: make separate function for this
// 	valueStr := fmt.Sprintf("%v", value)

// 	return valueStr, nil
// }

// GetValueServ получение значения метрики
func (s *Service) GetValueServ(metric models.Metric) (string, error) {
    // Проверка метрики
    if err := validateMetric(metric); err != nil {
        return "", err
    }

	value, err := s.Storage.GetValue(metric)
	if err != nil {
		log.Printf("failed to get value: %v", err)
		return "", err
	}

	// TODO: make sparate function for this
	valueStr := fmt.Sprintf("%v", value)

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
		strValue, ok := metric.Value.(string)
		if !ok {
			log.Println("metricValue must be a string for gauge type")
			return models.NewHTTPError(http.StatusBadRequest, "metricValue must be a string for gauge type")
		}
		value, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			log.Printf("invalid gauge value: %v", err)
			return models.NewHTTPError(http.StatusBadRequest, "invalid gauge value")
		}
		s.Storage.UpdateMetric(models.Metric{
			Type:  metric.Type,
			Name:  metric.Name,
			Value: value,
		})
	case "counter":
		// Получение старого значения счетчика
		counterVal, err := s.GetValueServ(metric)
        fmt.Println("Смотрим что нам возвращается2:", counterVal, "Ошибка:", err)
		if err != nil {
			if errors.Is(err, models.ErrMetricNotFound) {
				counterVal = "0"
			} else {
				return err
			}
		}

		counterInt, err := strconv.Atoi(counterVal)
		if err != nil {
			log.Printf("failed to convert value to int: %v", err)
			return models.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to convert value to int: %v", err))
		}
		// }

		strValue, ok := metric.Value.(string)
		if !ok {
			log.Println("metricValue must be a string for counter type")
			return models.NewHTTPError(http.StatusBadRequest, "metricValue must be a string for counter type")
		}
		newValue, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			log.Printf("invalid counter value: %v", err)
			return models.NewHTTPError(http.StatusBadRequest, "invalid counter value")
		}

		// Добавление старого значения к новому
		totalValue := newValue + int64(counterInt)
		s.Storage.UpdateMetric(models.Metric{
			Type:  metric.Type,
			Name:  metric.Name,
			Value: totalValue,
		})
	default:
		log.Printf("unknown metric type: %s", metric.Type)
		return models.NewHTTPError(http.StatusBadRequest, "unknown metric type")
	}

	return nil
}

// validateMetric проверяет метрику на наличие ошибок
func validateMetric(metric models.Metric) error {
	if metric.Type == "" || metric.Value == "" {
		log.Println("metric cannot be empty")
		return models.NewHTTPError(http.StatusBadRequest, "metricType cannot be empty")
	}

	if metric.Name == "" {
		log.Println("metricName cannot be empty")
		return models.NewHTTPError(http.StatusNotFound, "metricName cannot be empty")
	}

	return nil
}
