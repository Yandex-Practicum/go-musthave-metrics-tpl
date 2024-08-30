package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger структура для логгера
type Logger struct {
	ZapLogger   *zap.Logger
	AtomicLevel zap.AtomicLevel
}

// NewLogger создает новый экземпляр Logger
func NewLogger(level string, logFile string) (*Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	atomicLevel := zap.NewAtomicLevelAt(zapLevel)

	config := zap.Config{
		Level:       atomicLevel,
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout", logFile},
		ErrorOutputPaths: []string{"stderr"},
	}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{ZapLogger: zapLogger, AtomicLevel: atomicLevel}, nil
}

// Info логирует информационные сообщения
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.ZapLogger.Info(msg, fields...)
}

// Error логирует сообщения об ошибках
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.ZapLogger.Error(msg, fields...)
}

// Debug логирует отладочные сообщения
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.ZapLogger.Debug(msg, fields...)
}

// Warn логирует предупреждения
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.ZapLogger.Warn(msg, fields...)
}

// Sync синхронизирует логгер
func (l *Logger) Sync() {
	l.ZapLogger.Sync()
}
