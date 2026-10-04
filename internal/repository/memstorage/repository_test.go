package memstorage

import (
	"errors"
	"testing"
	"yandex-go-musthave-metrics-tpl/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	storage := NewMemStorage()
	testVal := 1.1
	tableTest := []struct {
		name     string
		metric   model.Metrics
		expected error
	}{
		{
			"add",
			model.Metrics{
				ID:    "Test",
				MType: model.Gauge,
				Delta: nil,
				Value: &testVal,
				Hash:  "",
			},
			nil,
		},
	}
	for _, tt := range tableTest {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.Add(t.Context(), tt.metric)
			if err != nil {
				assert.Equal(t, tt.expected, err)
				return
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	testVal := 1.1
	tableTest := []struct {
		name     string
		setup    func(m *MemStorage)
		metricID string
		expected float64
		err      error
	}{
		{
			"getValue",
			func(m *MemStorage) {
				m.data["Test"] = model.Metrics{
					ID:    "Test",
					MType: model.Gauge,
					Delta: nil,
					Value: &testVal,
					Hash:  "",
				}
			},
			"Test",
			1.1,
			nil,
		},
		{
			"getValueError",
			nil,
			"Test",
			0,
			errors.New("metric not found"),
		},
	}
	for _, tt := range tableTest {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			if tt.setup != nil {
				tt.setup(storage)
			}
			val, err := storage.GetValue(t.Context(), tt.metricID)
			if err != nil {
				assert.Equal(t, tt.err, err)
				return
			}
			assert.Equal(t, tt.expected, *val)
		})
	}
}

func TestGetAll(t *testing.T) {
	testVal := 1.1
	tableTest := []struct {
		name     string
		setup    func(m *MemStorage)
		expected []model.Metrics
		err      error
	}{
		{
			"getValue",
			func(m *MemStorage) {
				m.data["Test1"] = model.Metrics{
					ID:    "Test1",
					MType: model.Gauge,
					Delta: nil,
					Value: &testVal,
					Hash:  "",
				}
				m.data["Test2"] = model.Metrics{
					ID:    "Test2",
					MType: model.Counter,
					Delta: nil,
					Value: &testVal,
					Hash:  "",
				}
			},
			[]model.Metrics{
				model.Metrics{
					ID:    "Test1",
					MType: model.Gauge,
					Delta: nil,
					Value: &testVal,
					Hash:  "",
				},
				model.Metrics{
					ID:    "Test2",
					MType: model.Counter,
					Delta: nil,
					Value: &testVal,
					Hash:  "",
				},
			},
			nil,
		},
		{
			"getValueError",
			nil,
			nil,
			errors.New("no metrics found"),
		},
	}
	for _, tt := range tableTest {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			if tt.setup != nil {
				tt.setup(storage)
			}
			metrics, err := storage.GetAll(t.Context())
			if err != nil {
				assert.Equal(t, tt.err, err)
				return
			}
			assert.Equal(t, tt.expected, metrics)
		})
	}
}
