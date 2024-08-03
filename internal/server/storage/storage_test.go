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
        input         models.Metric
        expectedValue interface{}
        expectedError error
    }{
        {
            name: "MetrixStatistic - empty storage",
            setup: func() *Storage {
                return New()
            },
            method:        "MetrixStatistic",
            expectedValue: map[string]interface{}{},
            expectedError: nil,
        },
        {
            name: "UpdateMetric - add gauge metric",
            setup: func() *Storage {
                return New()
            },
            method: "UpdateMetric",
            input: models.Metric{
                Type:  "gauge",
                Name:  "testGauge",
                Value: 123.45,
            },
            expectedValue: nil,
            expectedError: nil,
        },
        {
            name: "GetValue - existing gauge metric",
            setup: func() *Storage {
                s := New()
                s.UpdateMetric(models.Metric{
                    Type:  "gauge",
                    Name:  "testGauge",
                    Value: 123.45,
                })
                return s
            },
            method: "GetValue",
            input: models.Metric{
                Type: "gauge",
                Name: "testGauge",
            },
            expectedValue: 123.45,
            expectedError: nil,
        },
        {
            name: "GetValue - non-existing metric",
            setup: func() *Storage {
                return New()
            },
            method: "GetValue",
            input: models.Metric{
                Type: "gauge",
                Name: "nonExisting",
            },
            expectedValue: "",
            expectedError: ErrMetricNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := tt.setup()

            switch tt.method {
            case "MetrixStatistic":
                value, err := s.MetrixStatistic()
                if err != tt.expectedError {
                    t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
                }
                if !equal(value, tt.expectedValue) {
                    t.Errorf("expected value: %v, got: %v", tt.expectedValue, value)
                }
            case "UpdateMetric":
                err := s.UpdateMetric(tt.input)
                if err != tt.expectedError {
                    t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
                }
            case "GetValue":
                value, err := s.GetValue(tt.input)
                if err != tt.expectedError {
                    t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
                }
                if value != tt.expectedValue {
                    t.Errorf("expected value: %v, got: %v", tt.expectedValue, value)
                }
            }
        })
    }
}

// equal - вспомогательная функция для сравнения значений
func equal(a, b interface{}) bool {
    switch a := a.(type) {
    case map[string]interface{}:
        b, ok := b.(map[string]interface{})
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