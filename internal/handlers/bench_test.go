package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

func BenchmarkListHandler(b *testing.B) {
	repo := storage.NewMemoryStorage()
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		repo.UpdateGauge(ctx, "gauge"+strconv.Itoa(i), float64(i)*1.5)
		repo.UpdateCounter(ctx, "counter"+strconv.Itoa(i), int64(i))
	}
	h := ListHandler(repo)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}
}
