package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestInitialize проверяет установку корректного уровня и ошибку при
// нераспознанном уровне логирования.
func TestInitialize(t *testing.T) {
	if err := Initialize("info"); err != nil {
		t.Fatalf("Initialize(info) вернул ошибку: %v", err)
	}
	if err := Initialize("error"); err != nil {
		t.Fatalf("Initialize(error) вернул ошибку: %v", err)
	}
	if err := Initialize("не-уровень"); err == nil {
		t.Fatal("ожидалась ошибка при некорректном уровне логирования")
	}
}

// TestRequestLogger проверяет, что middleware вызывает следующий обработчик
// и не меняет ответ, оборачивая ResponseWriter.
func TestRequestLogger(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Fatalf("запись ответа: %v", err)
		}
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	RequestLogger(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("следующий обработчик не был вызван")
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("ожидался статус 202, получен %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("тело ответа = %q, ожидалось %q", rec.Body.String(), "ok")
	}
}
