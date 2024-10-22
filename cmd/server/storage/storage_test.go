package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var StorageConfig = MemStorageConfig{
	StoreInterval:   300,
	FileStoragePath: "storage.json",
	Restore:         true,
}

func TestMemStorageSetAndGetGauge(t *testing.T) {
	s := NewMemStorage(StorageConfig)
	s.SetGauge("temperature", 23.5)

	value, exists := s.GetGauge("temperature")
	assert.True(t, exists)
	assert.Equal(t, 23.5, value)
}
func TestMemStorage_IncrementCounter(t *testing.T) {
	s := NewMemStorage(StorageConfig)
	s.IncrementCounter("hits", 10)
	s.IncrementCounter("hits", 5)

	value, exists := s.GetCounter("hits")
	assert.True(t, exists)
	assert.Equal(t, int64(15), value)
}
