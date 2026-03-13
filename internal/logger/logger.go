package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var once sync.Once

// Init инициализирует глобальный zerolog с выводом в ротационный файл logs/app.log
func Init() {
	once.Do(func() {
		// Создаем папку logs при необходимости
		err := os.MkdirAll("logs", 0o755)
		if err != nil {
			_, _ = os.Stderr.WriteString("Failed to create logs directory: " + err.Error() + "\n")
		}

		logFile := &lumberjack.Logger{
			Filename:   filepath.Join("logs", "app.log"),
			MaxSize:    100, // Мегабайты
			MaxBackups: 7,   // Количество резервных файлов
			MaxAge:     30,  // Максимальный возраст в днях
			Compress:   true,
		}

		multi := io.MultiWriter(logFile)

		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = zerolog.New(multi).With().Timestamp().Logger()

		log.Logger.Info().Msg("Logger initialized and writing to logs/app.log")
	})
}

// GetLogger возвращает глобальный zerolog.Logger
func GetLogger() zerolog.Logger {
	return log.Logger
}
