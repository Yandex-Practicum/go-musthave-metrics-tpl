package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/vova4o/yandexadv/internal/models"
	"github.com/vova4o/yandexadv/internal/server/flags"
	"github.com/vova4o/yandexadv/package/logger"
	"go.uber.org/zap"
)

// Storage структура для хранилища
type Storage struct {
	FileStorage *os.File
	Encoder     *json.Encoder
	MemStorage  map[string]models.Metrics
	mu          sync.Mutex
}

// // Loggerer интерфейс для логгера
// type Loggerer interface {
// 	Info(string, ...zap.Field)
// 	Error(string, ...zap.Field)
// }

// New создание нового хранилища
func New() *Storage {
	return &Storage{
		MemStorage: make(map[string]models.Metrics),
	}
}

// StartFileStorageLogic запуск логики хранения данных в файле
func StartFileStorageLogic(config *flags.Config, s *Storage, logger *logger.Logger) {
	if config.FileStoragePath != "" {
		err := s.OpenFile(config.FileStoragePath)
		if err != nil {
			logger.Error("Failed to open file: %v", zap.Error(err))
		}
	} else {
		logger.Info("File storage is not specified")
	}

	if config.Restore {
		err := s.LoadMemStorageFromFile()
		if err != nil {
			logger.Error("Failed to restore data from file: %v", zap.Error(err))
		}
	}

	go func() {
		for {
			time.Sleep(time.Duration(config.StoreInterval) * time.Second)
			s.SaveMemStorageToFile()
		}
	}()
}

// OpenFile открытие файла для хранения данных
func (s *Storage) OpenFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	encoder := json.NewEncoder(file)

	s.FileStorage = file
	s.Encoder = encoder

	return nil
}

// CloseFile закрытие файла
func (s *Storage) CloseFile() error {
	return s.FileStorage.Close()
}

// SaveMemStorageToFile сохранение данных из памяти в файл
func (s *Storage) SaveMemStorageToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Очистка файла
	if err := s.FileStorage.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	// Установка указателя файла в начало
	if _, err := s.FileStorage.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Запись данных из памяти в файл
	for _, metric := range s.MemStorage {
		if err := s.Encoder.Encode(metric); err != nil {
			return fmt.Errorf("failed to encode metric: %w", err)
		}
	}

	return nil
}

// LoadMemStorageFromFile загрузка данных из файла в память
func (s *Storage) LoadMemStorageFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Установка указателя файла в начало
	if _, err := s.FileStorage.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Создание декодера для чтения данных из файла
	decoder := json.NewDecoder(s.FileStorage)

	// Чтение данных из файла
	var metric models.Metrics
	for {
		if err := decoder.Decode(&metric); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode metric: %w", err)
		}

		s.MemStorage[metric.ID] = metric
	}

	return nil
}

// MetrixStatistic получение статистики метрик
func (s *Storage) MetrixStatistic() (map[string]models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var metrics = make(map[string]models.Metrics)

	for metricType, metricValues := range s.MemStorage {
		metrics[metricType] = metricValues
	}

	return metrics, nil
}

// UpdateMetric обновление метрики
func (s *Storage) UpdateMetric(metric models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.MemStorage[metric.ID] = metric
	return nil
}

// GetValue получение значения метрики по ID метрики
// возвращает значение метрики и ошибку
// возвращает значение не указателем, а значением
func (s *Storage) GetValue(metric models.Metrics) (*models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if val, ok := s.MemStorage[metric.ID]; ok {
		return &val, nil
	}

	return nil, models.ErrMetricNotFound
}
