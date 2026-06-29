package agent

import (
	"testing"
)

func TestCollectGauges(t *testing.T) {
	gauges := CollectGauges()

	expectedNames := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc", "RandomValue",
	}

	if len(gauges) != len(expectedNames) {
		t.Errorf("CollectGauges() : ожидали %d метрик,получили %d", len(expectedNames), len(gauges))
	}

	names := make(map[string]bool)
	for _, g := range gauges {
		names[g.Name] = true

	}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("CollectGauges(): отсутствует метрика %q", name)
		}
	}

}
