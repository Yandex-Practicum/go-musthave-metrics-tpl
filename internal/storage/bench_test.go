package storage

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	models "github.com/bluegopher/go-musthave-metrics-tpl/internal/model"
)

// makeBatch формирует пакет из n gauge- и n counter-метрик.
func makeBatch(n int) []models.Metrics {
	batch := make([]models.Metrics, 0, 2*n)
	for i := 0; i < n; i++ {
		v := float64(i) * 1.5
		d := int64(i)
		batch = append(batch,
			models.Metrics{ID: "gauge" + strconv.Itoa(i), MType: "gauge", Value: &v},
			models.Metrics{ID: "counter" + strconv.Itoa(i), MType: "counter", Delta: &d},
		)
	}
	return batch
}

// populate наполняет хранилище n gauge и n counter значениями.
func populate(s *MemoryStorage, n int) {
	ctx := context.Background()
	for i := 0; i < n; i++ {
		s.UpdateGauge(ctx, "gauge"+strconv.Itoa(i), float64(i)*1.5)
		s.UpdateCounter(ctx, "counter"+strconv.Itoa(i), int64(i))
	}
}

func BenchmarkUpdateBatchSync(b *testing.B) {
	batch := makeBatch(15)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewMemoryStorage()
		s.SetSyncFile(filepath.Join(b.TempDir(), "sync.json"))
		if err := s.UpdateBatch(ctx, batch); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateBatchMemory(b *testing.B) {
	batch := makeBatch(15)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewMemoryStorage()
		if err := s.UpdateBatch(ctx, batch); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSaveToFile(b *testing.B) {
	s := NewMemoryStorage()
	populate(s, 100)
	file := filepath.Join(b.TempDir(), "save.json")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := SaveToFile(s, file); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetAllGauges(b *testing.B) {
	s := NewMemoryStorage()
	populate(s, 100)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.GetAllGauges(ctx)
	}
}
