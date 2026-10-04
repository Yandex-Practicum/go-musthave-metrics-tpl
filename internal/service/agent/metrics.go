package agent

import (
	"math/rand"
	"runtime"
	"yandex-go-musthave-metrics-tpl/internal/model"
)

func GetMetrics() map[string]model.GaugeT {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]model.GaugeT{
		"Alloc":         model.GaugeT(memStats.Alloc),
		"BuckHashSys":   model.GaugeT(memStats.BuckHashSys),
		"Frees":         model.GaugeT(memStats.Frees),
		"GCCPUFraction": model.GaugeT(memStats.GCCPUFraction),
		"GCSys":         model.GaugeT(memStats.GCSys),
		"HeapAlloc":     model.GaugeT(memStats.HeapAlloc),
		"HeapIdle":      model.GaugeT(memStats.HeapIdle),
		"HeapInuse":     model.GaugeT(memStats.HeapInuse),
		"HeapObjects":   model.GaugeT(memStats.HeapObjects),
		"HeapReleased":  model.GaugeT(memStats.HeapReleased),
		"HeapSys":       model.GaugeT(memStats.HeapSys),
		"LastGC":        model.GaugeT(memStats.LastGC),
		"Lookups":       model.GaugeT(memStats.Lookups),
		"MCacheInuse":   model.GaugeT(memStats.MCacheInuse),
		"MCacheSys":     model.GaugeT(memStats.MCacheSys),
		"MSpanInuse":    model.GaugeT(memStats.MSpanInuse),
		"MSpanSys":      model.GaugeT(memStats.MSpanSys),
		"Mallocs":       model.GaugeT(memStats.Mallocs),
		"NextGC":        model.GaugeT(memStats.NextGC),
		"NumForcedGC":   model.GaugeT(memStats.NumForcedGC),
		"NumGC":         model.GaugeT(memStats.NumGC),
		"OtherSys":      model.GaugeT(memStats.OtherSys),
		"PauseTotalNs":  model.GaugeT(memStats.PauseTotalNs),
		"StackInuse":    model.GaugeT(memStats.StackInuse),
		"StackSys":      model.GaugeT(memStats.StackSys),
		"Sys":           model.GaugeT(memStats.Sys),
		"TotalAlloc":    model.GaugeT(memStats.TotalAlloc),
		"RandomValue":   model.GaugeT(rand.Float64()),
	}
}
