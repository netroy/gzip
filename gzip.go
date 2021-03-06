package gzip

import (
	"compress/gzip"
	"github.com/go-martini/martini"
	"net/http"
	"strings"
)

const (
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentLength   = "Content-Length"
	HeaderContentType     = "Content-Type"
	HeaderVary            = "Vary"
)

var serveGzip = func(w http.ResponseWriter, r *http.Request, c martini.Context) {
	if !strings.Contains(r.Header.Get(HeaderAcceptEncoding), "gzip") {
		return
	}

	headers := w.Header()
	headers.Set(HeaderContentEncoding, "gzip")
	headers.Set(HeaderVary, HeaderAcceptEncoding)

	gz := gzip.NewWriter(w)
	defer gz.Close()

	gzw := &gzipResponseWriter{gz, w.(martini.ResponseWriter), 0}
	c.MapTo(gzw, (*http.ResponseWriter)(nil))

	c.Next()

	// delete content length after we know we have been written to
	gzw.Header().Del("Content-Length")
}

// All returns a Handler that adds gzip compression to all requests
func All() martini.Handler {
	return serveGzip
}

type gzipResponseWriter struct {
	w *gzip.Writer
	martini.ResponseWriter
	status int
}

func (grw *gzipResponseWriter) WriteHeader(status int) {
	grw.status = status
}

func (grw gzipResponseWriter) Write(p []byte) (int, error) {
	rw := grw.ResponseWriter
	if !rw.Written() {
		status := grw.status
		if status == 0 {
			status = http.StatusOK
		}
		headers := grw.Header()
		if headers != nil && len(headers.Get(HeaderContentType)) == 0 {
			headers.Set(HeaderContentType, http.DetectContentType(p))
		}
		rw.WriteHeader(status)
	}
	return grw.w.Write(p)
}
