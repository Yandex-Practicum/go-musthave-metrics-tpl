package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	logFile := "test.log"
	defer os.Remove(logFile)

	logger, err := NewLogger("debug", logFile)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestLoggerInfo(t *testing.T) {
	logFile := "test.log"
	defer os.Remove(logFile)

	logger, err := NewLogger("info", logFile)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Info("This is an info message")
	logger.Sync()

	// Проверка содержимого лог-файла
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "This is an info message")
}

func TestLoggerError(t *testing.T) {
	logFile := "test.log"
	defer os.Remove(logFile)

	logger, err := NewLogger("error", logFile)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Error("This is an error message")
	logger.Sync()

	// Проверка содержимого лог-файла
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "This is an error message")
}

func TestLoggerDebug(t *testing.T) {
	logFile := "test.log"
	defer os.Remove(logFile)

	logger, err := NewLogger("debug", logFile)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Debug("This is a debug message")
	logger.Sync()

	// Проверка содержимого лог-файла
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "This is a debug message")
}

func TestLoggerWarn(t *testing.T) {
	logFile := "test.log"
	defer os.Remove(logFile)

	logger, err := NewLogger("warn", logFile)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger.Warn("This is a warn message")
	logger.Sync()

	// Проверка содержимого лог-файла
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "This is a warn message")
}
