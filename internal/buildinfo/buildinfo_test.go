package buildinfo

import (
	"bytes"
	"testing"
)

// TestPrint проверяет, что Print выводит переданные значения без изменений.
func TestPrint(t *testing.T) {
	tests := []struct {
		name                  string
		version, date, commit string
		want                  string
	}{
		{
			name:    "все значения заданы",
			version: "v1.0.0", date: "2026-07-08", commit: "abc123",
			want: "Build version: v1.0.0\nBuild date: 2026-07-08\nBuild commit: abc123\n",
		},
		{
			name:    "значения по умолчанию",
			version: "N/A", date: "N/A", commit: "N/A",
			want: "Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Print(&buf, tt.version, tt.date, tt.commit)
			if got := buf.String(); got != tt.want {
				t.Fatalf("Print() = %q, ожидалось %q", got, tt.want)
			}
		})
	}
}
