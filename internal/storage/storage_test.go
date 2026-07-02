package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_GaugeCounter(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStorage()

	require.NoError(t, s.UpdateGauge(ctx, "cpu", 3.14))
	v, ok := s.GetGauge(ctx, "cpu")
	assert.True(t, ok)
	assert.Equal(t, 3.14, v)

	require.NoError(t, s.UpdateCounter(ctx, "hits", 5))
	require.NoError(t, s.UpdateCounter(ctx, "hits", 7))
	c, ok := s.GetCounter(ctx, "hits")
	assert.True(t, ok)
	assert.Equal(t, int64(12), c)

	_, ok = s.GetGauge(ctx, "missing")
	assert.False(t, ok)
	_, ok = s.GetCounter(ctx, "missing")
	assert.False(t, ok)
}

func TestMemoryStorage_GetAll(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStorage()
	require.NoError(t, s.UpdateGauge(ctx, "g1", 1.0))
	require.NoError(t, s.UpdateGauge(ctx, "g2", 2.0))
	require.NoError(t, s.UpdateCounter(ctx, "c1", 10))

	gauges := s.GetAllGauges(ctx)
	assert.Len(t, gauges, 2)
	assert.Equal(t, 1.0, gauges["g1"])
	assert.Equal(t, 2.0, gauges["g2"])

	counters := s.GetAllCounters(ctx)
	assert.Len(t, counters, 1)
	assert.Equal(t, int64(10), counters["c1"])
}

func TestMemoryStorage_UpdateBatch(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStorage()

	g := 12.5
	d := int64(3)
	batch := []models.Metrics{
		{ID: "Alloc", MType: "gauge", Value: &g},
		{ID: "Poll", MType: "counter", Delta: &d},
		{ID: "Poll", MType: "counter", Delta: &d},
	}
	require.NoError(t, s.UpdateBatch(ctx, batch))

	v, ok := s.GetGauge(ctx, "Alloc")
	assert.True(t, ok)
	assert.Equal(t, 12.5, v)

	c, ok := s.GetCounter(ctx, "Poll")
	assert.True(t, ok)
	assert.Equal(t, int64(6), c)
}

func TestMemoryStorage_UpdateBatchSync(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStorage()
	path := filepath.Join(t.TempDir(), "sync.json")
	s.SetSyncFile(path)

	g := 1.0
	require.NoError(t, s.UpdateBatch(ctx, []models.Metrics{{ID: "g", MType: "gauge", Value: &g}}))

	_, err := os.Stat(path)
	require.NoError(t, err)
}

func TestSaveAndLoadFile(t *testing.T) {
	ctx := context.Background()
	src := NewMemoryStorage()
	require.NoError(t, src.UpdateGauge(ctx, "cpu", 9.9))
	require.NoError(t, src.UpdateCounter(ctx, "hits", 42))

	path := filepath.Join(t.TempDir(), "metrics.json")
	require.NoError(t, SaveToFile(src, path))

	dst := NewMemoryStorage()
	require.NoError(t, LoadFromFile(dst, path))

	v, ok := dst.GetGauge(ctx, "cpu")
	assert.True(t, ok)
	assert.Equal(t, 9.9, v)
	c, ok := dst.GetCounter(ctx, "hits")
	assert.True(t, ok)
	assert.Equal(t, int64(42), c)
}

func TestLoadFromFile_Missing(t *testing.T) {
	dst := NewMemoryStorage()
	err := LoadFromFile(dst, filepath.Join(t.TempDir(), "nope.json"))
	assert.Error(t, err)
}
