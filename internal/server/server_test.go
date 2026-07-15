package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/audit"
	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

// TestNew проверяет, что конструктор сохраняет переданные зависимости.
func TestNew(t *testing.T) {
	repo := storage.NewMemoryStorage()
	pub := audit.NewPublisher()
	defer pub.Close()

	s := New(":8080", repo, nil, "secret", pub, true, nil)
	if s.addr != ":8080" || s.hashKey != "secret" || !s.enablePprof {
		t.Fatalf("New сохранил зависимости некорректно: %+v", s)
	}
	if s.repo == nil || s.auditPub == nil {
		t.Fatal("New не сохранил repo/auditPub")
	}
}

// TestBuildRouter проверяет, что роутер регистрирует эндпоинты метрик и
// корректно обрабатывает запросы обновления и чтения.
func TestBuildRouter(t *testing.T) {
	repo := storage.NewMemoryStorage()
	pub := audit.NewPublisher()
	defer pub.Close()

	s := New(":0", repo, nil, "", pub, true, nil)
	ts := httptest.NewServer(s.buildRouter())
	defer ts.Close()

	// Обновляем gauge-метрику.
	resp, err := http.Post(ts.URL+"/update/gauge/Temp/36.6", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: ожидался 200, получен %d", resp.StatusCode)
	}

	// Читаем её обратно.
	resp, err = http.Get(ts.URL + "/value/gauge/Temp")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("value: ожидался 200, получен %d", resp.StatusCode)
	}

	// Корневая страница со списком метрик.
	resp, err = http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list: ожидался 200, получен %d", resp.StatusCode)
	}
}
