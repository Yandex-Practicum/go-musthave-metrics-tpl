package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vova4o/yandexadv/internal/models"
)

// MockStorager - мок для хранилища
type MockStorager struct {
    mock.Mock
}

func (m *MockStorager) UpdateMetric(metric models.Metrics) error {
    args := m.Called(metric)
    return args.Error(0)
}

func (m *MockStorager) GetValue(metric models.Metrics) (*models.Metrics, error) {
    args := m.Called(metric)
    return args.Get(0).(*models.Metrics), args.Error(1)
}

func (m *MockStorager) MetrixStatistic() (map[string]models.Metrics, error) {
    args := m.Called()
    return args.Get(0).(map[string]models.Metrics), args.Error(1)
}

func TestUpdateServ(t *testing.T) {
    mockStorage := new(MockStorager)
    service := New(mockStorage)

    t.Run("Update gauge metric", func(t *testing.T) {
        metric := models.Metric{
            Type:  "gauge",
            Name:  "test_metric_gauge",
            Value: "123.45",
        }

        valueFloat := 123.45
        mockStorage.On("UpdateMetric", models.Metrics{
            MType: metric.Type,
            ID:    metric.Name,
            Value: &valueFloat,
        }).Return(nil)

        err := service.UpdateServ(metric)
        assert.NoError(t, err)
        mockStorage.AssertExpectations(t)
    })

    t.Run("Update counter metric", func(t *testing.T) {
        metric := models.Metric{
            Type:  "counter",
            Name:  "test_metric_counter",
            Value: "678",
        }

        valueInt := int64(678)
        existingValueInt := int64(678)
        updatedValueInt := existingValueInt + valueInt

        mockStorage.On("GetValue", models.Metrics{
            MType: metric.Type,
            ID:    metric.Name,
        }).Return(&models.Metrics{Delta: &existingValueInt}, nil)

        mockStorage.On("UpdateMetric", models.Metrics{
            MType: metric.Type,
            ID:    metric.Name,
            Delta: &updatedValueInt,
        }).Return(nil)

        err := service.UpdateServ(metric)
        assert.NoError(t, err)
        mockStorage.AssertExpectations(t)
    })
}

func TestGetValueServ(t *testing.T) {
    mockStorage := new(MockStorager)
    service := New(mockStorage)

    t.Run("Get gauge metric value", func(t *testing.T) {
        metric := models.Metrics{
            MType: "gauge",
            ID:    "test_metric_gauge",
        }

        valueFloat := 123.45
        mockStorage.On("GetValue", metric).Return(&models.Metrics{Value: &valueFloat}, nil)

        value, err := service.GetValueServ(metric)
        assert.NoError(t, err)
        assert.Equal(t, "123.45", value)
        mockStorage.AssertExpectations(t)
    })

    t.Run("Get counter metric value", func(t *testing.T) {
        metric := models.Metrics{
            MType: "counter",
            ID:    "test_metric_counter",
        }

        valueInt := int64(678)
        mockStorage.On("GetValue", metric).Return(&models.Metrics{Delta: &valueInt}, nil)

        value, err := service.GetValueServ(metric)
        assert.NoError(t, err)
        assert.Equal(t, "678", value)
        mockStorage.AssertExpectations(t)
    })
}

func TestUpdateServJSON(t *testing.T) {
    mockStorage := new(MockStorager)
    service := New(mockStorage)

    t.Run("Update gauge metric JSON", func(t *testing.T) {
        metric := models.Metrics{
            MType: "gauge",
            ID:    "test_metric_gauge",
            Value: new(float64),
        }
        *metric.Value = 123.45

        mockStorage.On("UpdateMetric", metric).Return(nil)

        err := service.UpdateServJSON(metric)
        assert.NoError(t, err)
        mockStorage.AssertExpectations(t)
    })

    t.Run("Update counter metric JSON", func(t *testing.T) {
        metric := models.Metrics{
            MType: "counter",
            ID:    "test_metric_counter",
            Delta: new(int64),
        }
        *metric.Delta = 678

        mockStorage.On("GetValue", models.Metrics{
            MType: metric.MType,
            ID:    metric.ID,
        }).Return(&models.Metrics{Delta: new(int64)}, nil)

        mockStorage.On("UpdateMetric", models.Metrics{
            MType: metric.MType,
            ID:    metric.ID,
            Delta: metric.Delta,
        }).Return(nil)

        err := service.UpdateServJSON(metric)
        assert.NoError(t, err)
        mockStorage.AssertExpectations(t)
    })
}

func TestGetValueServJSON(t *testing.T) {
    mockStorage := new(MockStorager)
    service := New(mockStorage)

    t.Run("Get gauge metric value JSON", func(t *testing.T) {
        metric := models.Metrics{
            MType: "gauge",
            ID:    "test_metric_gauge",
        }

        valueFloat := 123.45
        mockStorage.On("GetValue", metric).Return(&models.Metrics{Value: &valueFloat}, nil)

        result, err := service.GetValueServJSON(metric)
        assert.NoError(t, err)
        assert.Equal(t, &valueFloat, result.Value)
        mockStorage.AssertExpectations(t)
    })

    t.Run("Get counter metric value JSON", func(t *testing.T) {
        metric := models.Metrics{
            MType: "counter",
            ID:    "test_metric_counter",
        }

        valueInt := int64(678)
        mockStorage.On("GetValue", metric).Return(&models.Metrics{Delta: &valueInt}, nil)

        result, err := service.GetValueServJSON(metric)
        assert.NoError(t, err)
        assert.Equal(t, &valueInt, result.Delta)
        mockStorage.AssertExpectations(t)
    })
}

func TestMetrixStatistic(t *testing.T) {
    mockStorage := new(MockStorager)
    service := New(mockStorage)

    t.Run("Get metrics statistics", func(t *testing.T) {
        metrics := map[string]models.Metrics{
            "test_metric_gauge": {
                MType: "gauge",
                ID:    "test_metric_gauge",
                Value: new(float64),
            },
            "test_metric_counter": {
                MType: "counter",
                ID:    "test_metric_counter",
                Delta: new(int64),
            },
        }
        *metrics["test_metric_gauge"].Value = 123.45
        *metrics["test_metric_counter"].Delta = 678

        mockStorage.On("MetrixStatistic").Return(metrics, nil)

        tmpl, result, err := service.MetrixStatistic()
        assert.NoError(t, err)
        assert.NotNil(t, tmpl)
        assert.Equal(t, metrics, result)
        mockStorage.AssertExpectations(t)
    })
}