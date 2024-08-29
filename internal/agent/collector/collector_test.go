package collector

import (
	"testing"
)

func TestCollectMetrics(t *testing.T) {
	tests := []struct {
		name      string
		pollCount int64
	}{
		{"Test with pollCount 0", 0},
		{"Test with pollCount 1", 1},
		{"Test with pollCount 100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollectMetrics(tt.pollCount)
			if len(got) == 0 {
				t.Errorf("CollectMetrics() returned empty slice")
			}

			// Check for specific metric values
			for _, metric := range got {
				switch metric.ID {
				case "PollCount":
					if v := metric.Delta; v != nil {
						if *v != tt.pollCount {
							t.Errorf("Expected PollCount to be %v, got %v", tt.pollCount, *v)
						}
					} else {
						t.Errorf("PollCount metric is nil")
					}
				case "RandomValue":
					if v := metric.Value; v != nil {
						if *v < 0 || *v > 1 {
							t.Errorf("Expected RandomValue to be between 0 and 1, got %v", *v)
						}
					} else {
						t.Errorf("RandomValue metric is nil")
					}
				default:
					// Check other metrics that use Value
					if metric.MType == "gauge" {
						if metric.Value == nil {
							t.Errorf("Expected Value for metric %v to be non-nil", metric.ID)
						}
					} else if metric.MType == "counter" {
						if metric.Delta == nil {
							t.Errorf("Expected Delta for metric %v to be non-nil", metric.ID)
						}
					}
				}
			}
		})
	}
}
