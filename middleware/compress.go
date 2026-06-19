package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

// compressResponseWriter interset for gzip and brotli compression
type compressResponseWriter struct {
	*StatusHTTP
	writer io.Writer
}

func (cw *compressResponseWriter) Write(b []byte) (int, error) {

	// detect Contetn-Type header and set it if not present
	if cw.Header().Get("Content-Type") == "" {
		cw.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return cw.writer.Write(b)
}

func (cw *compressResponseWriter) Flush() {

	if flusher, ok := cw.writer.(interface{ Flush() error }); ok {
		_ = flusher.Flush()
	}

	if flusher, ok := cw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Compress middleware for gzip and brotli
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// log.Println("8. Compress middleware fire!")

		// Skip WebSocket connections. They use their own compression mechanism.
		if req.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, req)
			return
		}

		acceptEncoding := req.Header.Get("Accept-Encoding")

		// 1. Brotli priority
		if strings.Contains(acceptEncoding, "br") {

			w.Header().Set("Content-Encoding", "br")
			w.Header().Add("Vary", "Accept-Encoding")

			// DefaultCompression (6)
			brWriter := brotli.NewWriterLevel(w, brotli.DefaultCompression)
			defer brWriter.Close()

			cw := &compressResponseWriter{
				StatusHTTP: NewStatusHTTP(w),
				writer:     brWriter,
			}
			next.ServeHTTP(cw, req)
			return
		}

		// 2. Gzip priority
		if strings.Contains(acceptEncoding, "gzip") {

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")

			// DefaultCompression (-1)
			gzWriter, _ := gzip.NewWriterLevel(w, gzip.DefaultCompression)
			defer gzWriter.Close()

			cw := &compressResponseWriter{
				StatusHTTP: NewStatusHTTP(w),
				writer:     gzWriter,
			}
			next.ServeHTTP(cw, req)
			return
		}

		// 3. if client does not support the compression, then return original response
		next.ServeHTTP(w, req)
	})
}
