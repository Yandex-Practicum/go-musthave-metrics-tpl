package metrics

import (
	"testing"
)

func TestMetricCreation(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metric
		expected string
	}{
		{
			name: "Gauge metric",
			metric: Metric{
				Name:  "metric1",
				Type:  "gauge",
				Value: 1.23,
			},
			expected: "metric1",
		},
		{
			name: "Counter metric",
			metric: Metric{
				Name:  "metric2",
				Type:  "counter",
				Value: 10,
			},
			expected: "metric2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metric.Name != tt.expected {
				t.Errorf("Expected metric name %s, but got %s", tt.expected, tt.metric.Name)
			}
		})
	}
}

func TestMetricValue(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metric
		expected interface{}
	}{
		{
			name: "Gauge metric value",
			metric: Metric{
				Name:  "metric1",
				Type:  "gauge",
				Value: 1.23,
			},
			expected: 1.23,
		},
		{
			name: "Counter metric value",
			metric: Metric{
				Name:  "metric2",
				Type:  "counter",
				Value: 10,
			},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metric.Value != tt.expected {
				t.Errorf("Expected metric value %v, but got %v", tt.expected, tt.metric.Value)
			}
		})
	}
}