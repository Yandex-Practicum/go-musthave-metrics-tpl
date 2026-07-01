package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// FileSink — приёмник аудита, дописывающий события в конец файла,
// по одному событию на строку (формат JSON Lines).
// Файл открывается в NewFileSink на всё время жизни приёмника и должен
// быть корректно закрыт через Close (например, в graceful shutdown сервиса),
// чтобы не терять последние записи и вернуть файловый дескриптор ОС.
type FileSink struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileSink создаёт файловый приёмник аудита и открывает файл в
// режиме дозаписи. Если файл не существует, он будет создан.
func NewFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileSink{file: f}, nil
}

// Notify добавляет событие в конец файла новой строкой в формате JSON.
// Сериализация и подготовка буфера выполняются вне мьютекса, чтобы
// критическая секция содержала только сам системный вызов записи.
func (s *FileSink) Notify(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	buf := append(data, '\n')

	s.mu.Lock()
	_, err = s.file.Write(buf)
	s.mu.Unlock()
	return err
}

// Close закрывает файл-приёмник. Безопасно вызывать многократно и на nil-приёмнике.
func (s *FileSink) Close() error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == nil {
		return nil
	}
	err := s.file.Close()
	s.file = nil
	return err
}
