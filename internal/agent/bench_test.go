package agent

import "testing"

func BenchmarkCollectGauges(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = CollectGauges()
	}
}
