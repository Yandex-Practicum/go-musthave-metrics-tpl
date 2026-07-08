// Package buildinfo форматирует и выводит сведения о сборке бинарника:
// версию, дату и коммит. Значения задаются при сборке через
// -ldflags "-X main.buildVersion=...".
package buildinfo

import (
	"fmt"
	"io"
)

// Print выводит в w сведения о сборке. Пустые значения заменяются на "N/A".
func Print(w io.Writer, version, date, commit string) {
	fmt.Fprintf(w, "Build version: %s\n", orNA(version))
	fmt.Fprintf(w, "Build date: %s\n", orNA(date))
	fmt.Fprintf(w, "Build commit: %s\n", orNA(commit))
}

// orNA возвращает s или "N/A", если строка пуста.
func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
