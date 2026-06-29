package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// FileSink — приёмник аудита, дописывающий события в конец файла,
// по одному событию на строку (формат JSON Lines).
type FileSink struct {
	mu   sync.Mutex
	path string
}

// NewFileSink создаёт файловый приёмник аудита.
func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

// Notify добавляет событие в конец файла новой строкой в формате JSON.
func (s *FileSink) Notify(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(append(data, '\n'))
	return err
}
