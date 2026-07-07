// Package agent реализует сбор метрик рантайма и их периодическую
// отправку на сервер.
package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// GaugeMetric — метрика типа gauge (float64).
type GaugeMetric struct {
	Name  string
	Value float64
}

// CounterMetric — метрика типа counter (int64).
type CounterMetric struct {
	Name  string
	Delta int64
}

// CollectGauges собирает gauge-метрики из runtime.
func CollectGauges() []GaugeMetric {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	gauges := []GaugeMetric{
		{"Alloc", float64(m.Alloc)},
		{"BuckHashSys", float64(m.BuckHashSys)},
		{"Frees", float64(m.Frees)},
		{"GCCPUFraction", m.GCCPUFraction},
		{"GCSys", float64(m.GCSys)},
		{"HeapAlloc", float64(m.HeapAlloc)},
		{"HeapIdle", float64(m.HeapIdle)},
		{"HeapInuse", float64(m.HeapInuse)},
		{"HeapObjects", float64(m.HeapObjects)},
		{"HeapReleased", float64(m.HeapReleased)},
		{"HeapSys", float64(m.HeapSys)},
		{"LastGC", float64(m.LastGC)},
		{"Lookups", float64(m.Lookups)},
		{"MCacheInuse", float64(m.MCacheInuse)},
		{"MCacheSys", float64(m.MCacheSys)},
		{"MSpanInuse", float64(m.MSpanInuse)},
		{"MSpanSys", float64(m.MSpanSys)},
		{"Mallocs", float64(m.Mallocs)},
		{"NextGC", float64(m.NextGC)},
		{"NumForcedGC", float64(m.NumForcedGC)},
		{"NumGC", float64(m.NumGC)},
		{"OtherSys", float64(m.OtherSys)},
		{"PauseTotalNs", float64(m.PauseTotalNs)},
		{"StackInuse", float64(m.StackInuse)},
		{"StackSys", float64(m.StackSys)},
		{"Sys", float64(m.Sys)},
		{"TotalAlloc", float64(m.TotalAlloc)},
		{"RandomValue", rand.Float64()},
	}

	return gauges
}

func CollectPSUtilMetrics() []GaugeMetric {
	var metrics []GaugeMetric

	v, err := mem.VirtualMemory()
	if err != nil {
		metrics = append(metrics,
			GaugeMetric{Name: "TotalMemory", Value: float64(v.Total)},
			GaugeMetric{Name: "FreeMemory", Value: float64(v.Free)},
		)
	}
	cpuPercent, err := cpu.Percent(0, true)
	if err == nil {
		for i, pct := range cpuPercent {
			name := fmt.Sprintf("CPUutilization%d", i+1)
			metrics = append(metrics, GaugeMetric{Name: name, Value: pct})
		}
	}

	return metrics
}
