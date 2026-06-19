package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogMiddleware(t *testing.T) {

	// Arrange
	var buf bytes.Buffer
	logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://test.test/log", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")

	mockLogHandler := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}

	rr := httptest.NewRecorder()

	// Act
	handler := Log(http.HandlerFunc(mockLogHandler))
	handler.ServeHTTP(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if got := rr.Body.String(); got != "test response" {
		t.Fatalf("unexpected body: %q", got)
	}

	if buf.Len() == 0 {
		t.Error("Empty log output!")
	}
}
