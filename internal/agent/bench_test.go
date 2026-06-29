package agent

import "testing"

func BenchmarkCollectGauges(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CollectGauges()
	}
}
