package metrics

import (
	"testing"
)

// toFloat64Pointer преобразует значение float64 в указатель на float64
func toFloat64Pointer(value float64) *float64 {
	return &value
}

// toInt64Pointer преобразует значение int64 в указатель на int64
func toInt64Pointer(value int64) *int64 {
	return &value
}

func TestMetricCreation(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metrics
		expected string
	}{
		{
			name: "Gauge metric",
			metric: Metrics{
				ID:    "metric1",
				MType: "gauge",
				Value: toFloat64Pointer(1.23),
			},
			expected: "metric1",
		},
		{
			name: "Counter metric",
			metric: Metrics{
				ID:    "metric2",
				MType: "counter",
				Delta: toInt64Pointer(10),
			},
			expected: "metric2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metric.ID != tt.expected {
				t.Errorf("Expected metric name %s, but got %s", tt.expected, tt.metric.ID)
			}
		})
	}
}

func TestMetricValue(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metrics
		expected interface{}
	}{
		{
			name: "Gauge metric value",
			metric: Metrics{
				ID:    "metric1",
				MType: "gauge",
				Value: toFloat64Pointer(1.23),
			},
			expected: 1.23,
		},
		{
			name: "Counter metric value",
			metric: Metrics{
				ID:    "metric2",
				MType: "counter",
				Delta: toInt64Pointer(10),
			},
			expected: int64(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metric.MType == "gauge" {
				if *tt.metric.Value != tt.expected {
					t.Errorf("Expected metric value %v, but got %v", tt.expected, *tt.metric.Value)
				}
			} else if tt.metric.MType == "counter" {
				if *tt.metric.Delta != tt.expected {
					t.Errorf("Expected metric value %v, but got %v", tt.expected, *tt.metric.Delta)
				}
			}
		})
	}
}
