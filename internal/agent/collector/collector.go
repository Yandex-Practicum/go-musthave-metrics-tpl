package collector

import (
	"math/rand"
	"runtime"

	"github.com/vova4o/yandexadv/internal/agent/metrics"
)

// toFloat64Pointer преобразует значение float64 в указатель на float64
func toFloat64Pointer(value float64) *float64 {
	return &value
}

// CollectMetrics собирает метрики и возвращает их
func CollectMetrics(pollCount int64) []metrics.Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return []metrics.Metrics{
		{ID: "Alloc", MType: "gauge", Value: toFloat64Pointer(float64(m.Alloc))},
		{ID: "BuckHashSys", MType: "gauge", Value: toFloat64Pointer(float64(m.BuckHashSys))},
		{ID: "Frees", MType: "gauge", Value: toFloat64Pointer(float64(m.Frees))},
		{ID: "GCCPUFraction", MType: "gauge", Value: toFloat64Pointer(m.GCCPUFraction)},
		{ID: "GCSys", MType: "gauge", Value: toFloat64Pointer(float64(m.GCSys))},
		{ID: "HeapAlloc", MType: "gauge", Value: toFloat64Pointer(float64(m.HeapAlloc))},
		{ID: "HeapIdle", MType: "gauge", Value: toFloat64Pointer(float64(m.HeapIdle))},
		{ID: "HeapInuse", MType: "gauge", Value: toFloat64Pointer(float64(m.HeapInuse))},
		{ID: "HeapObjects", MType: "gauge", Value: toFloat64Pointer(float64(m.HeapObjects))},
		{ID: "HeapReleased", MType: "gauge", Value: toFloat64Pointer(float64(m.HeapReleased))},
		{ID: "HeapSys", MType: "gauge", Value: toFloat64Pointer(float64(m.HeapSys))},
		{ID: "LastGC", MType: "gauge", Value: toFloat64Pointer(float64(m.LastGC))},
		{ID: "Lookups", MType: "gauge", Value: toFloat64Pointer(float64(m.Lookups))},
		{ID: "MCacheInuse", MType: "gauge", Value: toFloat64Pointer(float64(m.MCacheInuse))},
		{ID: "MCacheSys", MType: "gauge", Value: toFloat64Pointer(float64(m.MCacheSys))},
		{ID: "MSpanInuse", MType: "gauge", Value: toFloat64Pointer(float64(m.MSpanInuse))},
		{ID: "MSpanSys", MType: "gauge", Value: toFloat64Pointer(float64(m.MSpanSys))},
		{ID: "Mallocs", MType: "gauge", Value: toFloat64Pointer(float64(m.Mallocs))},
		{ID: "NextGC", MType: "gauge", Value: toFloat64Pointer(float64(m.NextGC))},
		{ID: "NumForcedGC", MType: "gauge", Value: toFloat64Pointer(float64(m.NumForcedGC))},
		{ID: "NumGC", MType: "gauge", Value: toFloat64Pointer(float64(m.NumGC))},
		{ID: "OtherSys", MType: "gauge", Value: toFloat64Pointer(float64(m.OtherSys))},
		{ID: "PauseTotalNs", MType: "gauge", Value: toFloat64Pointer(float64(m.PauseTotalNs))},
		{ID: "StackInuse", MType: "gauge", Value: toFloat64Pointer(float64(m.StackInuse))},
		{ID: "StackSys", MType: "gauge", Value: toFloat64Pointer(float64(m.StackSys))},
		{ID: "Sys", MType: "gauge", Value: toFloat64Pointer(float64(m.Sys))},
		{ID: "TotalAlloc", MType: "gauge", Value: toFloat64Pointer(float64(m.TotalAlloc))},
		{ID: "PollCount", MType: "counter", Delta: &pollCount},
		{ID: "RandomValue", MType: "gauge", Value: toFloat64Pointer(rand.Float64())},
	}
}
