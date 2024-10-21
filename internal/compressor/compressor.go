package compressor

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request URL: %s, Content-Encoding: %s, Accept-Encoding: %s", r.URL, r.Header.Get("Content-Encoding"), r.Header.Get("Accept-Encoding"))

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Unable to read gzip data", http.StatusBadRequest)
				return
			}
			defer func() {
				err := gzipReader.Close()
				if err != nil {
					http.Error(w, "Unable to close gzip data", http.StatusBadRequest)
				}
			}()

			r.Body = io.NopCloser(gzipReader)
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer func() {
				err := gz.Close()
				if err != nil {
					http.Error(w, "Unable to close gzip data", http.StatusBadRequest)
				}
			}()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			gzw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
			next.ServeHTTP(gzw, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}
