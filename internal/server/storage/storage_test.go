package storage

import (
	"testing"

	"github.com/vova4o/yandexadv/internal/models"
)

func TestStorage(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *Storage
		method        string
		input         models.Metrics
		expectedValue *float64
		expectedError error
	}{
		{
			name: "GetValue - existing gauge metric",
			setup: func() *Storage {
				s := New()
				s.UpdateMetric(models.Metrics{
					MType: "gauge",
					ID:    "testGauge",
					Value: func() *float64 {
						v := 123.45
						return &v
					}(),
				})
				return s
			},
			method: "GetValue",
			input: models.Metrics{
				MType: "gauge",
				ID:    "testGauge",
			},
			expectedValue: func() *float64 {
				v := 123.45
				return &v
			}(),
			expectedError: nil,
		},
		{
			name: "GetValue - non-existing metric",
			setup: func() *Storage {
				return New()
			},
			method: "GetValue",
			input: models.Metrics{
				MType: "gauge",
				ID:    "nonExisting",
			},
			expectedValue: nil,
			expectedError: models.ErrMetricNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()

			switch tt.method {
			case "GetValue":
				metric, err := s.GetValue(tt.input)
				if err != tt.expectedError {
					t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
				}
				if metric == nil || metric.Value == nil {
					if tt.expectedValue != nil {
						t.Errorf("expected value: %v, got: %v", tt.expectedValue, metric)
					}
				} else if *metric.Value != *tt.expectedValue {
					t.Errorf("expected value: %v, got: %v", *tt.expectedValue, *metric.Value)
				}
			}
		})
	}
}

// equal - вспомогательная функция для сравнения значений
func equal(a, b interface{}) bool {
	switch a := a.(type) {
	case map[string]models.Metrics:
		b, ok := b.(map[string]models.Metrics)
		if !ok {
			return false
		}
		if len(a) != len(b) {
			return false
		}
		for k, v := range a {
			if !equal(v, b[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
