package storage

import (
	"testing"
)

func TestStorage_Update(t *testing.T) {
    tests := []struct {
        name        string
        metricType  string
        metricName  string
        metricValue interface{}
        wantErr     bool
    }{
        {
            name:        "Update existing metric",
            metricType:  "counter",
            metricName:  "requests",
            metricValue: 10,
            wantErr:     false,
        },
        {
            name:        "Add new metric type",
            metricType:  "gauge",
            metricName:  "temperature",
            metricValue: 23.5,
            wantErr:     false,
        },
        {
            name:        "Update existing metric with new value",
            metricType:  "counter",
            metricName:  "requests",
            metricValue: 20,
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := New()
            if err := s.Update(tt.metricType, tt.metricName, tt.metricValue); (err != nil) != tt.wantErr {
                t.Errorf("Storage.Update() error = %v, wantErr %v", err, tt.wantErr)
            }

            if got := s.MemStorage[tt.metricType][tt.metricName]; got != tt.metricValue {
                t.Errorf("Storage.Update() = %v, want %v", got, tt.metricValue)
            }
        })
    }
}