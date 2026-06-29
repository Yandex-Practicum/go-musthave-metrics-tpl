package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/service"
)

func ListHandler(srv service.MetricsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gauges := srv.GetAllGauges(r.Context())
		counters := srv.GetAllCounters(r.Context())

		var sb strings.Builder
		sb.WriteString("<html><body><ul>")

		for name, value := range gauges {
			sb.WriteString("<li>gauge ")
			sb.WriteString(name)
			sb.WriteString(" = ")
			sb.WriteString(strconv.FormatFloat(value, 'g', -1, 64))
			sb.WriteString("</li>")
		}

		for name, value := range counters {
			sb.WriteString("<li>counter ")
			sb.WriteString(name)
			sb.WriteString(" = ")
			sb.WriteString(strconv.FormatInt(value, 10))
			sb.WriteString("</li>")
		}
		sb.WriteString("</ul></body></html>")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, sb.String())
	}
}
