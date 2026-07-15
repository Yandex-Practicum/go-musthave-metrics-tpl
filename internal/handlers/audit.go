package handlers

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/audit"
)

// publishAudit формирует событие аудита по обработанным метрикам и
// ставит его в очередь приёмников. Асинхронность и ограничение
// параллельных отправок обеспечивает Publisher.Notify через worker pool,
// поэтому здесь не нужно запускать новую горутину на каждый запрос.
func publishAudit(pub *audit.Publisher, r *http.Request, metrics []string) {
	if !pub.Enabled() || len(metrics) == 0 {
		return
	}
	pub.Notify(audit.Event{
		TS:        time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: clientIP(r),
	})
}

// clientIP определяет IP входящего запроса: сначала по заголовкам
// прокси (X-Real-IP, X-Forwarded-For), затем по RemoteAddr.
func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
