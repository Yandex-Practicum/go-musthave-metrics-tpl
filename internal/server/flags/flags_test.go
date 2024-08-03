package flags

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	// Сбросим viper перед каждым тестом
	viper.Reset()

	// Сбросим флаги перед каждым тестом
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	// Установим переменные окружения
	os.Setenv("ADDRESS", "127.0.0.1:9090")
	os.Setenv("STORE_INTERVAL", "15")
	os.Setenv("FILE_STORAGE_PATH", "test")
	os.Setenv("RESTORE", "true")

	// Создадим новую конфигурацию
	config := NewConfig()

	// Проверим значения из переменных окружения
	assert.Equal(t, "127.0.0.1:9090", config.ServerAddress)
	assert.Equal(t, 15, config.StoreInterval)
	assert.Equal(t, "test", config.FileStoragePath)
	assert.Equal(t, true, config.Restore)

	// Очистим переменные окружения
	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")
}
