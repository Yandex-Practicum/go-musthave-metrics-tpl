package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bluegopher/go-musthave-metrics-tpl/internal/storage"
)

func ListHandler(repo storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gauges := repo.GetAllGauges()
		counters := repo.GetAllCounters()

		var sb strings.Builder
		sb.WriteString("<html><body><ul>")

		for name, value := range gauges {
			sb.WriteString(fmt.Sprintf("<li>gauge %s = %g</li>", name, value))
		}

		for name, value := range counters {
			sb.WriteString(fmt.Sprintf("<li>counter %s = %d</li>", name, value))
		}
		sb.WriteString("</ul></body></html>")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, sb.String())
	}
}
