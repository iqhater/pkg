package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

// JSON output
var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Log middleware handler shows network data log info
func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// log.Println("4. Log middleware fire!")

		t := time.Now()

		sr := NewStatusHTTP(w)
		next.ServeHTTP(sr, req)

		statusCode := sr.StatusCode

		layout := "02.01.2006 15:04:05"

		logger.Info("HTTP request",
			"time", t.Format(layout),
			"status_code", statusCode,
			"status_text", http.StatusText(statusCode),
			"duration", time.Since(t),
			"remote_addr", req.RemoteAddr,
			"method", req.Method,
			"content_type", req.Header.Get("Content-Type"),
			"content_length", req.Header.Get("Content-Length"),
			"url", req.URL.String(),
			"request_id", IDFromContext(req.Context()),
		)
	})
}
