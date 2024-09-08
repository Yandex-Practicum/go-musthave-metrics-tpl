package service

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vova4o/yandexadv/internal/models"
)

// MockStorager is a mock implementation of the Storager interface
type MockStorager struct {
	mock.Mock
}

func (m *MockStorager) UpdateBatch(metrics []models.Metrics) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func (m *MockStorager) UpdateMetric(metric models.Metrics) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *MockStorager) GetValue(metric models.Metrics) (*models.Metrics, error) {
	args := m.Called(metric)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Metrics), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorager) MetrixStatistic() (map[string]models.Metrics, error) {
	args := m.Called()
	return args.Get(0).(map[string]models.Metrics), args.Error(1)
}

func (m *MockStorager) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func TestUpdateServJSON(t *testing.T) {
	mockStorage := new(MockStorager)
	service := &Service{Storage: mockStorage}

	t.Run("Update gauge metric JSON", func(t *testing.T) {
		metric := &models.Metrics{
			MType: "gauge",
			ID:    "test_metric_gauge",
			Value: new(float64),
		}
		*metric.Value = 123.45

		mockStorage.On("UpdateMetric", *metric).Return(nil)

		err := service.UpdateServJSON(metric)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Update counter metric JSON", func(t *testing.T) {
		metric := &models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
			Delta: new(int64),
		}
		*metric.Delta = 678

		mockStorage.On("GetValue", models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
		}).Return(&models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
			Delta: new(int64),
		}, nil)

		mockStorage.On("UpdateMetric", mock.MatchedBy(func(m models.Metrics) bool {
			expectedValue := int64(678)
			return m.MType == "counter" && m.ID == "test_metric_counter" && *m.Delta == expectedValue
		})).Return(nil)

		err := service.UpdateServJSON(metric)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Unknown metric type", func(t *testing.T) {
		metric := &models.Metrics{
			MType: "unknown",
			ID:    "test_metric_unknown",
		}

		err := service.UpdateServJSON(metric)
		assert.Error(t, err)
		httpErr, ok := err.(*models.HTTPError)
		if ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Status)
		} else {
			t.Fatalf("expected *models.HTTPError, got %T", err)
		}
	})
}

func TestGetValueServJSON(t *testing.T) {
	mockStorage := new(MockStorager)
	service := &Service{Storage: mockStorage}

	t.Run("Get gauge metric JSON", func(t *testing.T) {
		metric := models.Metrics{
			MType: "gauge",
			ID:    "test_metric_gauge",
		}
		expectedValue := 123.45
		mockStorage.On("GetValue", metric).Return(&models.Metrics{
			MType: "gauge",
			ID:    "test_metric_gauge",
			Value: &expectedValue,
		}, nil)

		value, err := service.GetValueServJSON(metric)
		assert.NoError(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, expectedValue, *value.Value)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Get counter metric JSON", func(t *testing.T) {
		metric := models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
		}
		expectedDelta := int64(678)
		mockStorage.On("GetValue", metric).Return(&models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
			Delta: &expectedDelta,
		}, nil)

		value, err := service.GetValueServJSON(metric)
		assert.NoError(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, expectedDelta, *value.Delta)
		mockStorage.AssertExpectations(t)
	})
}

func TestMetrixStatistic(t *testing.T) {
	mockStorage := new(MockStorager)
	service := &Service{Storage: mockStorage}

	t.Run("Get metrics statistics", func(t *testing.T) {
		expectedMetrics := map[string]models.Metrics{
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
		*expectedMetrics["test_metric_gauge"].Value = 123.45
		*expectedMetrics["test_metric_counter"].Delta = 678

		mockStorage.On("MetrixStatistic").Return(expectedMetrics, nil)

		tmpl, metrics, err := service.MetrixStatistic()
		assert.NoError(t, err)
		assert.NotNil(t, tmpl)
		assert.Equal(t, expectedMetrics, metrics)
		mockStorage.AssertExpectations(t)
	})
}

func TestGetValueServ(t *testing.T) {
	mockStorage := new(MockStorager)
	service := &Service{Storage: mockStorage}

	t.Run("Get gauge metric", func(t *testing.T) {
		metric := models.Metrics{
			MType: "gauge",
			ID:    "test_metric_gauge",
		}
		expectedValue := 123.45
		mockStorage.On("GetValue", metric).Return(&models.Metrics{
			MType: "gauge",
			ID:    "test_metric_gauge",
			Value: &expectedValue,
		}, nil)

		value, err := service.GetValueServ(metric)
		assert.NoError(t, err)
		assert.Equal(t, strconv.FormatFloat(expectedValue, 'f', -1, 64), value)
		mockStorage.AssertExpectations(t)
	})

	t.Run("Get counter metric", func(t *testing.T) {
		metric := models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
		}
		expectedDelta := int64(678)
		mockStorage.On("GetValue", metric).Return(&models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
			Delta: &expectedDelta,
		}, nil)

		value, err := service.GetValueServ(metric)
		assert.NoError(t, err)
		assert.Equal(t, strconv.FormatInt(expectedDelta, 10), value)
		mockStorage.AssertExpectations(t)
	})
}

func TestUpdateServ(t *testing.T) {
	mockStorage := new(MockStorager)
	service := &Service{Storage: mockStorage}

	t.Run("Update gauge metric", func(t *testing.T) {
		metric := models.Metric{
			Type:  "gauge",
			Name:  "test_metric_gauge",
			Value: "123.45",
		}
		expectedValue := 123.45

		mockStorage.On("UpdateMetric", models.Metrics{
			MType: "gauge",
			ID:    "test_metric_gauge",
			Value: &expectedValue,
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
		expectedDelta := int64(678)

		mockStorage.On("GetValue", models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
		}).Return(&models.Metrics{
			MType: "counter",
			ID:    "test_metric_counter",
			Delta: new(int64),
		}, nil)

		mockStorage.On("UpdateMetric", mock.MatchedBy(func(m models.Metrics) bool {
			return m.MType == "counter" && m.ID == "test_metric_counter" && *m.Delta == expectedDelta
		})).Return(nil)

		err := service.UpdateServ(metric)
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})
}
