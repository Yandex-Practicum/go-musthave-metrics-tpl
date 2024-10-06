package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/vova4o/yandexadv/internal/models"
	"github.com/vova4o/yandexadv/internal/server/flags"
	"go.uber.org/zap"
)

// FileAndMemStorage структура для хранилища
type FileAndMemStorage struct {
	FileStorage *os.File
	Encoder     *json.Encoder
	MS          MemStorage
	mu          sync.Mutex
}

// NewFileStorage создание нового хранилища
func NewFileStorage() *FileAndMemStorage {
	return &FileAndMemStorage{
		MS: MemStorage{
			MemStorage: make(map[string]models.Metrics),
		},
	}
}

// SaveMemStorageToFile сохранение данных из памяти в файл
func (s *FileAndMemStorage) SaveMemStorageToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Очистка файла
	if err := s.FileStorage.Truncate(0); err != nil {
		log.Fatal(err)
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	// Установка указателя файла в начало
	if _, err := s.FileStorage.Seek(0, 0); err != nil {
		log.Fatal(err)
		return fmt.Errorf("failed to seek file: %w", err)
	}

	if err := s.Encoder.Encode(s.MS.MemStorage); err != nil {
		log.Fatal(err)
		return fmt.Errorf("failed to encode metrics: %w", err)
	}

	return nil
}

// LoadMemStorageFromFile загрузка данных из файла в память
func (s *FileAndMemStorage) LoadMemStorageFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Установка указателя файла в начало
	if _, err := s.FileStorage.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Создание декодера для чтения данных из файла
	decoder := json.NewDecoder(s.FileStorage)

	// Чтение данных из файла
	var metrics map[string]models.Metrics
	for {
		if err := decoder.Decode(&metrics); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode metric: %w", err)
		}

		s.MS.MemStorage = metrics
	}

	return nil
}

// StartFileStorageLogic запуск логики хранения данных в файле
func StartFileStorageLogic(config *flags.Config, s *FileAndMemStorage, logger Loggerer) {
	if config.FileStoragePath != "" {
		err := s.OpenFile(config.FileStoragePath)
		if err != nil {
			logger.Error("Failed to open file: %v", zap.Error(err))
		}
	} else {
		logger.Info("File storage is not specified")
		return
	}

	if config.Restore {
		err := s.LoadMemStorageFromFile()
		if err != nil {
			logger.Error("Failed to restore data from file: %v", zap.Error(err))
		}
	}

	go func() {
		for {
			interval := time.Duration(config.StoreInterval) * time.Second
			// if interval == 0 {
			// 	interval = 100 * time.Microsecond // Установите разумное значение по умолчанию
			// }
			time.Sleep(interval)
			s.SaveMemStorageToFile()
		}
	}()
}

// OpenFile открытие файла для хранения данных
func (s *FileAndMemStorage) OpenFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	encoder := json.NewEncoder(file)

	s.FileStorage = file
	s.Encoder = encoder

	return nil
}

// Stop закрытие файла
func (s *FileAndMemStorage) Stop() error {
	s.SaveMemStorageToFile()
	return s.FileStorage.Close()
}

// Ping проверка подключения к файлу
func (s *FileAndMemStorage) Ping() error {
	return nil
}

// UpdateMetric обновление метрики
func (s *FileAndMemStorage) UpdateMetric(metric models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.MS.MemStorage[metric.ID] = metric

	return nil
}

// GetValue получение значения метрики по ID метрики
func (s *FileAndMemStorage) GetValue(metric models.Metrics) (*models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if val, ok := s.MS.MemStorage[metric.ID]; ok {
		return &val, nil
	}

	return nil, models.ErrMetricNotFound
}

// MetrixStatistic получение статистики метрик
func (s *FileAndMemStorage) MetrixStatistic() (map[string]models.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var metrics = make(map[string]models.Metrics)

	for metricType, metricValues := range s.MS.MemStorage {
		metrics[metricType] = metricValues
	}

	return metrics, nil
}

// UpdateBatch обновление метрик по пакетно
func (s *FileAndMemStorage) UpdateBatch(metrics []models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range metrics {
		s.MS.MemStorage[metric.ID] = metric
	}

	return nil
}
