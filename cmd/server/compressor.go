package server

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type compressedResponseWriter struct {
	w   http.ResponseWriter
	gzw *gzip.Writer
}

func (c *compressedResponseWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressedResponseWriter) WriteHeader(i int) {
	c.w.WriteHeader(i)
}

func (c *compressedResponseWriter) Write(b []byte) (int, error) {
	return c.gzw.Write(b)
}

func compressor(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Accept-Encoding")

		if !strings.Contains(encoding, "gzip") {
			f(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/html") // default the content type
		w.Header().Del("Content-Length")

		ww := &compressedResponseWriter{w: w, gzw: gzip.NewWriter(w)}
		f(ww, r)
		ww.gzw.Close()
	}
}
