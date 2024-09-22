package service

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/vova4o/yandexadv/internal/models"
	"github.com/vova4o/yandexadv/package/logger"
	"go.uber.org/zap"
)

// Service структура для бизнес-логики
type Service struct {
	Storage Storager
	logger *logger.Logger
}

// Storager интерфейс для хранилища
type Storager interface {
	UpdateBatch(metrics []models.Metrics) error
	UpdateMetric(metric models.Metrics) error
	GetValue(metric models.Metrics) (*models.Metrics, error)
	MetrixStatistic() (map[string]models.Metrics, error)
	Ping() error
}

// New создание нового сервиса
func New(s Storager, logger *logger.Logger) *Service {
	return &Service{
		Storage: s,
		logger: logger,
	}
}

// UpdateBatchMetricsServ обновление метрик в формате JSON by batch
func (s *Service) UpdateBatchMetricsServ(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		log.Printf("Empty metrics")
		return models.NewHTTPError(http.StatusBadRequest, "Empty metrics")
	}
	s.logger.Info("Received POST JSON metrics for update", zap.Any("metrics", metrics))

	for _, metric := range metrics {
		err := s.UpdateServJSON(&metric)
		if err != nil {
			log.Printf("failed to update metric: %v", err)
			s.logger.Error("Failed to update metric", zap.Error(err))
			return err
		}
	}

	return nil
}

// PingDB проверка подключения к базе данных
func (s *Service) PingDB() error {
	return s.Storage.Ping()
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
	if value.Delta == nil && value.Value == nil {
		return nil, models.ErrMetricNotFound
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
			if errors.Is(err, models.ErrMetricNotFound) || errors.Is(err, sql.ErrNoRows) {
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
		err = s.Storage.UpdateMetric(models.Metrics{
			MType: metric.MType,
			ID:    metric.ID,
			Delta: &totalValue,
		})
		if err != nil {
			log.Printf("failed to update metric: %v", err)
			return err
		}
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

// // UpdateBatchMetricsServ обновление метрик в формате JSON by batch
// func (s *Service) UpdateBatchMetricsServ(metrics []models.Metrics) error {
//     if len(metrics) == 0 {
//         log.Printf("Empty metrics")
//         return models.NewHTTPError(http.StatusBadRequest, "Empty metrics")
//     }

//     // Создаем карту для хранения суммированных значений метрик
//     metricsMap := make(map[string]models.Metrics)

//     // Проверяем и суммируем метрики
//     for _, metric := range metrics {
//         if err := validateMetricJSON(&metric); err != nil {
//             return err
//         }

//         if existingMetric, exists := metricsMap[metric.ID]; exists {
//             if metric.MType == "counter" && existingMetric.MType == "counter" {
//                 if existingMetric.Delta != nil && metric.Delta != nil {
//                     *existingMetric.Delta += *metric.Delta
//                 }
//             } else {
//                 metricsMap[metric.ID] = metric
//             }
//         } else {
//             metricsMap[metric.ID] = metric
//         }
//     }

//     // Преобразуем карту обратно в срез
//     metricsToUpdate := make([]models.Metrics, 0, len(metricsMap))
//     for _, metric := range metricsMap {
//         metricsToUpdate = append(metricsToUpdate, metric)
//     }

//     // Обновляем метрики в хранилище
//     err := s.Storage.UpdateBatch(metricsToUpdate)
//     if err != nil {
//         log.Printf("failed to update metrics: %v", err)
//         return err
//     }

//     return nil
// }