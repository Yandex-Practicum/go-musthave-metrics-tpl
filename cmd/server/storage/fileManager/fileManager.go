package fileManager

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type StorageInterface interface {
	SetGauge(metricName string, value float64)
	IncrementCounter(metricName string, value int64)
	GetGauge(metricName string) (float64, bool)
	GetCounter(metricName string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}

type FileManager struct{}

func (fm *FileManager) SaveData(filePath string, storage StorageInterface) error {
	// Приведение storage к карте с сохранением состояния
	storageMap := map[string]interface{}{
		"gauges":   storage.GetAllGauges(),
		"counters": storage.GetAllCounters(),
	}
	fileData, err := json.Marshal(storageMap)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, fileData, 0644)
}

func (fm *FileManager) LoadData(filePath string, storage StorageInterface) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
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
