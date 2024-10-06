package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vova4o/yandexadv/internal/models"
	"github.com/vova4o/yandexadv/internal/server/storage"
)

func TestNewMemStorage(t *testing.T) {
	memStorage := storage.NewMemStorage()
	assert.NotNil(t, memStorage)
	assert.NotNil(t, memStorage.MemStorage)
}

func TestMemStorage_UpdateBatch(t *testing.T) {
	memStorage := storage.NewMemStorage()
	value1 := float64(10)
	value2 := float64(20)
	metrics := []models.Metrics{
		{ID: "metric1", Value: &value1},
		{ID: "metric2", Value: &value2},
	}

	err := memStorage.UpdateBatch(metrics)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(memStorage.MemStorage))
	assert.Equal(t, metrics[0], memStorage.MemStorage["metric1"])
	assert.Equal(t, metrics[1], memStorage.MemStorage["metric2"])
}

func TestMemStorage_UpdateMetric(t *testing.T) {
	memStorage := storage.NewMemStorage()
	value1 := float64(10)
	metric := models.Metrics{ID: "metric1", Value: &value1}

	err := memStorage.UpdateMetric(metric)
	assert.NoError(t, err)
	assert.Equal(t, metric, memStorage.MemStorage["metric1"])
}

func TestMemStorage_GetValue(t *testing.T) {
	memStorage := storage.NewMemStorage()
	value1 := float64(10)
	metric := models.Metrics{ID: "metric1", Value: &value1}
	memStorage.MemStorage[metric.ID] = metric

	val, err := memStorage.GetValue(metric)
	assert.NoError(t, err)
	assert.Equal(t, &metric, val)

	nonExistentMetric := models.Metrics{ID: "nonexistent"}
	val, err = memStorage.GetValue(nonExistentMetric)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestMemStorage_MetrixStatistic(t *testing.T) {
	memStorage := storage.NewMemStorage()
	val1 := float64(10)
	val2 := float64(20)
	metrics := []models.Metrics{
		{ID: "metric1", Value: &val1},
		{ID: "metric2", Value: &val2},
	}
	memStorage.UpdateBatch(metrics)

	stats, err := memStorage.MetrixStatistic()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(stats))
	assert.Equal(t, metrics[0], stats["metric1"])
	assert.Equal(t, metrics[1], stats["metric2"])
}

func TestMemStorage_Ping(t *testing.T) {
	memStorage := storage.NewMemStorage()
	err := memStorage.Ping()
	assert.NoError(t, err)
}

func TestMemStorage_Stop(t *testing.T) {
	memStorage := storage.NewMemStorage()
	err := memStorage.Stop()
	assert.NoError(t, err)
}
