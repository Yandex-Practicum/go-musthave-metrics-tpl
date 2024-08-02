package service

import (
	"testing"
)

type mockStorage struct {
	updateFunc func(metricType, metricName string, metricValue interface{}) error
}

func (m *mockStorage) Update(metricType, metricName string, metricValue interface{}) error {
	return m.updateFunc(metricType, metricName, metricValue)
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue interface{}
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Empty metricType",
			metricType:  "",
			metricName:  "requests",
			metricValue: "10",
			wantErr:     true,
			errMsg:      "metricType cannot be empty",
		},
		{
			name:        "Empty metricName",
			metricType:  "counter",
			metricName:  "",
			metricValue: "10",
			wantErr:     true,
			errMsg:      "metricName cannot be empty",
		},
		{
			name:        "Empty metricValue",
			metricType:  "counter",
			metricName:  "requests",
			metricValue: "",
			wantErr:     true,
			errMsg:      "metricValue cannot be nil",
		},
		{
			name:        "Invalid gauge value",
			metricType:  "gauge",
			metricName:  "temperature",
			metricValue: "invalid",
			wantErr:     true,
			errMsg:      "invalid gauge value",
		},
		{
			name:        "Invalid counter value",
			metricType:  "counter",
			metricName:  "requests",
			metricValue: "invalid",
			wantErr:     true,
			errMsg:      "invalid counter value",
		},
		{
			name:        "Unknown metric type",
			metricType:  "unknown",
			metricName:  "requests",
			metricValue: "10",
			wantErr:     true,
			errMsg:      "unknown metric type",
		},
		{
			name:        "Valid gauge update",
			metricType:  "gauge",
			metricName:  "temperature",
			metricValue: "23.5",
			wantErr:     false,
		},
		{
			name:        "Valid counter update",
			metricType:  "counter",
			metricName:  "requests",
			metricValue: "10",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mockStorage{
				updateFunc: func(metricType, metricName string, metricValue interface{}) error {
					return nil
				},
			}

			s := &Service{
				Storage: mockStorage,
			}

			err := s.Update(tt.metricType, tt.metricName, tt.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("Service.Update() error = %v, errMsg %v", err, tt.errMsg)
			}
		})
	}
}

