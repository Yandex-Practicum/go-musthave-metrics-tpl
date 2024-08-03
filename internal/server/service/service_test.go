package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/vova4o/yandexadv/internal/models"
)

// mockStorager - мок для интерфейса Storager
type mockStorager struct {
	UpdateMetricFunc    func(metric models.Metric) error
	GetValueFunc        func(metric models.Metric) (interface{}, error)
	MetrixStatisticFunc func() (map[string]interface{}, error)
}

func (m *mockStorager) UpdateMetric(metric models.Metric) error {
	if m.UpdateMetricFunc != nil {
		return m.UpdateMetricFunc(metric)
	}
	return nil
}

func (m *mockStorager) GetValue(metric models.Metric) (interface{}, error) {
	if m.GetValueFunc != nil {
		return m.GetValueFunc(metric)
	}
	return nil, nil
}

func (m *mockStorager) MetrixStatistic() (map[string]interface{}, error) {
	if m.MetrixStatisticFunc != nil {
		return m.MetrixStatisticFunc()
	}
	return nil, nil
}

func TestService_GetValue(t *testing.T) {
	mockStorage := &mockStorager{
		GetValueFunc: func(metric models.Metric) (interface{}, error) {
			if metric.Name == "test" {
				return "value", nil
			}
			return nil, errors.New("not found")
		},
	}

	s := New(mockStorage)

	value, err := s.GetValueServ(models.Metric{Type: "gauge", Name: "test"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "value" {
		t.Errorf("expected value 'value', got %v", value)
	}
}

func TestService_UpdateServ(t *testing.T) {
	tests := []struct {
		name          string
		metric        models.Metric
		mockStorage   *mockStorager
		expectedError error
	}{
		{
			name: "successful gauge update",
			metric: models.Metric{
				Type:  "gauge",
				Name:  "testGauge",
				Value: "123.45",
			},
			mockStorage: &mockStorager{
				UpdateMetricFunc: func(metric models.Metric) error {
					if metric.Type != "gauge" || metric.Name != "testGauge" || metric.Value != 123.45 {
						return errors.New("unexpected metric values")
					}
					return nil
				},
			},
			expectedError: nil,
		},
		// {
		//     name: "successful counter update",
		//     metric: models.Metric{
		//         Type:  "counter",
		//         Name:  "testCounter",
		//         Value: "10",
		//     },
		//     mockStorage: &mockStorager{
		//         GetValueFunc: func(metric models.Metric) (string, error) {
		//             return "5", nil
		//         },
		//         UpdateMetricFunc: func(metric models.Metric) error {
		//             if metric.Type != "counter" || metric.Name != "testCounter" || metric.Value != int64(15) {
		//                 return errors.New("unexpected metric values")
		//             }
		//             return nil
		//         },
		//     },
		//     expectedError: nil,
		// },
		{
			name: "validation error - empty type",
			metric: models.Metric{
				Type:  "",
				Name:  "testMetric",
				Value: "123",
			},
			mockStorage:   &mockStorager{},
			expectedError: models.NewHTTPError(http.StatusBadRequest, "metricType cannot be empty"),
		},
		{
			name: "validation error - empty name",
			metric: models.Metric{
				Type:  "gauge",
				Name:  "",
				Value: "123",
			},
			mockStorage:   &mockStorager{},
			expectedError: models.NewHTTPError(http.StatusNotFound, "metricName cannot be empty"),
		},
		// {
		//     name: "storage is nil",
		//     metric: models.Metric{
		//         Type:  "gauge",
		//         Name:  "testGauge",
		//         Value: "123.45",
		//     },
		//     mockStorage:   nil,
		//     expectedError: models.NewHTTPError(http.StatusInternalServerError, "storage cannot be nil"),
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{Storage: tt.mockStorage}
			err := s.UpdateServ(tt.metric)
			if err != nil && tt.expectedError == nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil && tt.expectedError != nil {
				t.Errorf("expected error: %v, got nil", tt.expectedError)
			}
			if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestValidateMetric(t *testing.T) {
	tests := []struct {
		name          string
		metric        models.Metric
		expectedError error
	}{
		{
			name: "empty type",
			metric: models.Metric{
				Type:  "",
				Name:  "testMetric",
				Value: "123",
			},
			expectedError: models.NewHTTPError(http.StatusBadRequest, "metricType cannot be empty"),
		},
		{
			name: "empty value",
			metric: models.Metric{
				Type:  "gauge",
				Name:  "testMetric",
				Value: "",
			},
			expectedError: models.NewHTTPError(http.StatusBadRequest, "metricType cannot be empty"),
		},
		{
			name: "empty name",
			metric: models.Metric{
				Type:  "gauge",
				Name:  "",
				Value: "123",
			},
			expectedError: models.NewHTTPError(http.StatusNotFound, "metricName cannot be empty"),
		},
		{
			name: "valid metric",
			metric: models.Metric{
				Type:  "gauge",
				Name:  "testMetric",
				Value: "123",
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetric(tt.metric)
			if err != nil && tt.expectedError == nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil && tt.expectedError != nil {
				t.Errorf("expected error: %v, got nil", tt.expectedError)
			}
			if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}
