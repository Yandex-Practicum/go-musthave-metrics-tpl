package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vova4o/yandexadv/package/logger"
	"go.uber.org/zap"
)

// Middleware структура для middleware
type Middleware struct {
	Logger *logger.Logger
}

// New создание нового middleware
func New(log *logger.Logger) *Middleware {
	return &Middleware{
		Logger: log,
	}
}

// GinZap возвращает middleware для логирования запросов с использованием zap
func (l Middleware) GinZap() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery

        c.Next()

        latency := time.Since(start)

        if raw != "" {
            path = path + "?" + raw
        }

        // Получение размера содержимого ответа
        contentLength := c.Writer.Header().Get("Content-Length")
        if contentLength == "" {
            contentLength = "0"
        }

        // Преобразование размера содержимого в int
        contentLengthInt, err := strconv.Atoi(contentLength)
        if err != nil {
            l.Logger.Error("failed to parse content length", zap.String("content_length", contentLength), zap.Error(err))
            contentLengthInt = 0 // или установите значение по умолчанию
        }

        // Получение и парсинг значения заголовка X-Response-Time
        latencyStr := c.Writer.Header().Get("X-Response-Time")
        var parsedLatency time.Duration
        if latencyStr != "" {
            parsedLatency, err = time.ParseDuration(latencyStr)
            if err != nil {
                l.Logger.Error("failed to parse latency", zap.String("latency", latencyStr), zap.Error(err))
                parsedLatency = 0 // или установите значение по умолчанию
            }
        }

        l.Logger.Info("incoming request",
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.Duration("latency", latency),
            zap.Int("status", c.Writer.Status()),
            zap.String("client_ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
            zap.Int("content_length", contentLengthInt),
            zap.Duration("parsed_latency", parsedLatency),
        )
    }
}