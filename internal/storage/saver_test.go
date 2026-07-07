package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRunSaver проверяет, что периодический сохранитель создаёт файл со
// снимком хранилища хотя бы за один тик.
func TestRunSaver(t *testing.T) {
	s := NewMemoryStorage()
	s.UpdateGauge(context.Background(), "Alloc", 1.5)

	file := filepath.Join(t.TempDir(), "snapshot.json")
	go RunSaver(s, file, 10*time.Millisecond)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("RunSaver не создал файл за отведённое время")
		default:
			if _, err := os.Stat(file); err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// TestSaveToFile_Error проверяет, что SaveToFile возвращает ошибку при
// невозможности создать файл (путь содержит несуществующий каталог).
func TestSaveToFile_Error(t *testing.T) {
	s := NewMemoryStorage()
	badPath := filepath.Join(t.TempDir(), "no-such-dir", "snapshot.json")
	if err := SaveToFile(s, badPath); err == nil {
		t.Fatal("ожидалась ошибка при создании файла в несуществующем каталоге")
	}
}
