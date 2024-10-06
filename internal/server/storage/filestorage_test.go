package storage_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vova4o/yandexadv/internal/models"
	"github.com/vova4o/yandexadv/internal/server/storage"
)

func TestNewFileStorage(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	assert.NotNil(t, fileStorage)
	assert.NotNil(t, fileStorage.MS.MemStorage)
}

func TestFileAndMemStorage_UpdateBatch(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	value1 := float64(10)
	value2 := float64(20)
	metrics := []models.Metrics{
		{ID: "metric1", Value: &value1},
		{ID: "metric2", Value: &value2},
	}

	err := fileStorage.UpdateBatch(metrics)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(fileStorage.MS.MemStorage))
	assert.Equal(t, metrics[0], fileStorage.MS.MemStorage["metric1"])
	assert.Equal(t, metrics[1], fileStorage.MS.MemStorage["metric2"])
}

func TestFileAndMemStorage_UpdateMetric(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	value := float64(10)
	metric := models.Metrics{ID: "metric1", Value: &value}

	err := fileStorage.UpdateMetric(metric)
	assert.NoError(t, err)
	assert.Equal(t, metric, fileStorage.MS.MemStorage["metric1"])
}

func TestFileAndMemStorage_GetValue(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	value := float64(10)
	metric := models.Metrics{ID: "metric1", Value: &value}
	fileStorage.MS.MemStorage[metric.ID] = metric

	val, err := fileStorage.GetValue(metric)
	assert.NoError(t, err)
	assert.Equal(t, &metric, val)

	nonExistentMetric := models.Metrics{ID: "nonexistent"}
	val, err = fileStorage.GetValue(nonExistentMetric)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestFileAndMemStorage_MetrixStatistic(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	value1 := float64(10)
	value2 := float64(20)
	metrics := []models.Metrics{
		{ID: "metric1", Value: &value1},
		{ID: "metric2", Value: &value2},
	}
	fileStorage.UpdateBatch(metrics)

	stats, err := fileStorage.MetrixStatistic()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(stats))
	assert.Equal(t, metrics[0], stats["metric1"])
	assert.Equal(t, metrics[1], stats["metric2"])
}

func TestFileAndMemStorage_Ping(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	err := fileStorage.Ping()
	assert.NoError(t, err)
}

func TestFileAndMemStorage_Stop(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	file, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	fileStorage.FileStorage = file
	fileStorage.Encoder = json.NewEncoder(file)

	err = fileStorage.Stop()
	assert.NoError(t, err)
}

func TestFileAndMemStorage_SaveMemStorageToFile(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	file, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	fileStorage.FileStorage = file
	fileStorage.Encoder = json.NewEncoder(file)

	value := float64(10)
	metric := models.Metrics{ID: "metric1", Value: &value}
	fileStorage.MS.MemStorage[metric.ID] = metric

	err = fileStorage.SaveMemStorageToFile()
	assert.NoError(t, err)

	// Проверка содержимого файла
	file.Seek(0, 0)
	decoder := json.NewDecoder(file)
	var metrics map[string]models.Metrics
	err = decoder.Decode(&metrics)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(metrics))
	assert.Equal(t, metric, metrics["metric1"])
}

func TestFileAndMemStorage_LoadMemStorageFromFile(t *testing.T) {
	fileStorage := storage.NewFileStorage()
	file, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	fileStorage.FileStorage = file
	fileStorage.Encoder = json.NewEncoder(file)

	value := float64(10)
	metric := models.Metrics{ID: "metric1", Value: &value}
	metrics := map[string]models.Metrics{
		metric.ID: metric,
	}

	// Запись данных в файл
	err = fileStorage.Encoder.Encode(metrics)
	assert.NoError(t, err)

	// Загрузка данных из файла
	err = fileStorage.LoadMemStorageFromFile()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(fileStorage.MS.MemStorage))
	assert.Equal(t, metric, fileStorage.MS.MemStorage["metric1"])
}

// func TestStartFileStorageLogic(t *testing.T) {
//     config := &flags.Config{
//         FileStoragePath: "/tmp/testfile",
//         StoreInterval:   1,
//         Restore:         true,
//     }
//     fileStorage := storage.NewFileStorage()
//     mockLogger := NewMockLogger()

//     // Настройка ожиданий для методов Info и Error
//     mockLogger.On("Info", mock.Anything, mock.Anything).Return()
//     mockLogger.On("Error", mock.Anything, mock.Anything).Return()

//     // Запуск логики хранения данных в файле
//     go storage.StartFileStorageLogic(config, fileStorage, mockLogger)

//     // Даем немного времени для выполнения горутины
//     time.Sleep(2 * time.Second)

//     // Проверка вызова методов
//     mockLogger.AssertExpectations(t)
// }
