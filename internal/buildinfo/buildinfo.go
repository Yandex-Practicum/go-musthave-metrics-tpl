// Package buildinfo форматирует и выводит сведения о сборке бинарника:
// версию, дату и коммит. Значения задаются при сборке через
// -ldflags "-X main.buildVersion=...".
package buildinfo

import (
	"fmt"
	"io"
)

// Print выводит в w сведения о сборке в том виде, в каком они переданы.
// Значения по умолчанию (например, "N/A") задаёт вызывающая сторона.
func Print(w io.Writer, version, date, commit string) {
	fmt.Fprintf(w, "Build version: %s\n", version)
	fmt.Fprintf(w, "Build date: %s\n", date)
	fmt.Fprintf(w, "Build commit: %s\n", commit)
}
