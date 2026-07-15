// Package middleware содержит HTTP-middleware: сжатие gzip и проверку
// подписи запросов.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// gzipWriterPool переиспользует gzip.Writer между запросами: их создание
// (compress/flate.NewWriter) — самый дорогой источник аллокаций под нагрузкой.
var gzipWriterPool = sync.Pool{
	New: func() any {
		gz, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return gz
	},
}

type gzipWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = gr
		}

		contentType := r.Header.Get("Accept")
		supportsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		compressible := strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html") ||
			contentType == ""

		if supportsGzip && compressible {
			gz := gzipWriterPool.Get().(*gzip.Writer)
			gz.Reset(w)
			defer func() {
				gz.Close()
				gzipWriterPool.Put(gz)
			}()
			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gzipWriter{ResponseWriter: w, writer: gz}, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
