package service

import (
	"context"
	"testing"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newService() MetricsService {
	return NewMetricsService(storage.NewMemoryStorage())
}

func TestService_GaugeCounter(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	require.NoError(t, svc.UpdateGauge(ctx, "cpu", 1.5))
	v, ok := svc.GetGauge(ctx, "cpu")
	assert.True(t, ok)
	assert.Equal(t, 1.5, v)

	require.NoError(t, svc.UpdateCounter(ctx, "hits", 4))
	c, ok := svc.GetCounter(ctx, "hits")
	assert.True(t, ok)
	assert.Equal(t, int64(4), c)
}

func TestService_GetAll(t *testing.T) {
	ctx := context.Background()
	svc := newService()
	require.NoError(t, svc.UpdateGauge(ctx, "g", 2.0))
	require.NoError(t, svc.UpdateCounter(ctx, "c", 9))

	assert.Equal(t, 2.0, svc.GetAllGauges(ctx)["g"])
	assert.Equal(t, int64(9), svc.GetAllCounters(ctx)["c"])
}

func TestService_UpdateBatch(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	g := 7.0
	d := int64(2)
	require.NoError(t, svc.UpdateBatch(ctx, []models.Metrics{
		{ID: "g", MType: "gauge", Value: &g},
		{ID: "c", MType: "counter", Delta: &d},
	}))

	v, ok := svc.GetGauge(ctx, "g")
	assert.True(t, ok)
	assert.Equal(t, 7.0, v)
	c, ok := svc.GetCounter(ctx, "c")
	assert.True(t, ok)
	assert.Equal(t, int64(2), c)
}
