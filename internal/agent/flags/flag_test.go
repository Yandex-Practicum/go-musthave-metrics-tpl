package flags

import (
	"os"
	"testing"
	"time"

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
	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("POLL_INTERVAL", "5")

	// Создадим новую конфигурацию
	config := NewConfig()

	// Проверим значения из переменных окружения
	assert.Equal(t, "127.0.0.1:9090", config.ServerAddress)
	assert.Equal(t, 15*time.Second, config.ReportInterval)
	assert.Equal(t, 5*time.Second, config.PollInterval)

	// Очистим переменные окружения
	os.Unsetenv("ADDRESS")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("POLL_INTERVAL")
}
