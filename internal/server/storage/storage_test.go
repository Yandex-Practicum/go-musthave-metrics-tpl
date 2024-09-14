package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vova4o/yandexadv/internal/server/flags"
	"github.com/vova4o/yandexadv/internal/server/storage"
	"go.uber.org/zap"
)

// MockLogger для тестирования
type MockLogger struct {
    mock.Mock
}

// Info логирует информационные сообщения
func (m *MockLogger) Info(msg string, fields ...zap.Field) {
    m.Called(msg, fields)
}

// Error логирует сообщения об ошибках
func (m *MockLogger) Error(msg string, fields ...zap.Field) {
    m.Called(msg, fields)
}

// NewMockLogger создает новый экземпляр MockLogger
func NewMockLogger() *MockLogger {
    return &MockLogger{}
}

func TestInit_NoStorageSelected(t *testing.T) {
    config := &flags.Config{}
    mockLogger := NewMockLogger()

    // Настройка ожиданий для методов Info и Error
    mockLogger.On("Error", "No storage selected using default: MemoryStorage", mock.Anything).Return()

    stor := storage.Init(config, mockLogger)
    assert.IsType(t, &storage.MemStorage{}, stor)

    // Проверка вызова методов
    mockLogger.AssertExpectations(t)
}


// func TestInit_DBStorageSelected(t *testing.T) {
//     config := &flags.Config{
//         DBDSN: "postgres://postgres:mypassword@localhost:5432/metrix?sslmode=disable",
//     }
//     mockLogger := NewMockLogger()

//     // Настройка ожиданий для методов Info и Error
//     mockLogger.On("Info", "Selected storage: DB", mock.Anything).Return()
//     mockLogger.On("Error", mock.Anything, mock.Anything).Return()

//     // Mock DBConnect to avoid actual DB connection
//     originalDBConnect := storage.DBConnect
//     defer func() { storage.DBConnect = originalDBConnect }()
//     storage.DBConnect = func(config *flags.Config, logger storage.Loggerer) (*storage.DBStorage, error) {
//         return &storage.DBStorage{}, nil
//     }

//     stor := storage.Init(config, mockLogger)
//     assert.IsType(t, &storage.DBStorage{}, stor)

//     // Проверка вызова методов
//     mockLogger.AssertExpectations(t)
// }

func TestInit_FileStorageSelected(t *testing.T) {
    config := &flags.Config{
        FileStoragePath: "/tmp/storage",
    }
    mockLogger := NewMockLogger()

    // Настройка ожиданий для методов Info и Error
    mockLogger.On("Info", "Selected storage: File", mock.Anything).Return()

    stor := storage.Init(config, mockLogger)
    assert.IsType(t, &storage.FileAndMemStorage{}, stor)

    // Проверка вызова методов
    mockLogger.AssertExpectations(t)
}