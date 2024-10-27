package storage

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type MemStorage struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

type MemStorageConfig struct {
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

func (m *MemStorage) SetGauge(metricName string, value float64) {
	m.Gauges[metricName] = value
}

func (m *MemStorage) IncrementCounter(metricName string, value int64) {
	m.Counters[metricName] += value
}

func (m *MemStorage) GetGauge(metricName string) (float64, bool) {
	value, exists := m.Gauges[metricName]
	return value, exists
}

func (m *MemStorage) GetCounter(metricName string) (int64, bool) {
	value, exists := m.Counters[metricName]
	return value, exists
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	return m.Gauges
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	return m.Counters
}

func NewMemStorage(config MemStorageConfig) *MemStorage {
	var storage = &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}

	if config.Restore {
		err := loadData(config.FileStoragePath, storage)
		if err != nil {
			log.Fatal(err)
		}
	}
	return storage
}

func loadData(filePath string, storage *MemStorage) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal("Failed to close file", err.Error())
		}
	}()

	if err != nil {
		log.Fatal("Can't open file.", err.Error())
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Can't read file.", err.Error())
	}
	if len(fileData) == 0 {
		log.Println("Storage file is empty, nothing to load.")
		return nil
	}

	err = json.Unmarshal(fileData, &storage)
	if err != nil {
		log.Fatal("Can't read json.", err.Error())
	}
	return nil
}

func (m *MemStorage) SaveData(filePath string) error {
	data := map[string]interface{}{
		"gauges":   m.Gauges,
		"counters": m.Counters,
	}

	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, fileData, 0644)
}
