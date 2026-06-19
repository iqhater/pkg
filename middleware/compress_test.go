package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

// mockHandler is a simple handler that returns a fixed response.
func mockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, this is a test response for compression!"))
}

func TestCompressBrotli(t *testing.T) {

	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "br")
	rec := httptest.NewRecorder()
	handler := Compress(http.HandlerFunc(mockHandler))

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Header().Get("Content-Encoding") != "br" {
		t.Errorf("expected Content-Encoding br, got %q", rec.Header().Get("Content-Encoding"))
	}

	// unpack brotli and check the content
	brReader := brotli.NewReader(rec.Body)
	decompressed, err := io.ReadAll(brReader)
	if err != nil {
		t.Fatalf("failed to decompress brotli: %v", err)
	}

	expected := "Hello, this is a test response for compression!"
	if string(decompressed) != expected {
		t.Errorf("expected body %q, got %q", expected, string(decompressed))
	}
}

func TestCompressGzip(t *testing.T) {

	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	handler := Compress(http.HandlerFunc(mockHandler))

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding gzip, got %q", rec.Header().Get("Content-Encoding"))
	}

	// unpack gzip and check the content
	gzReader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("failed to decompress gzip: %v", err)
	}

	expected := "Hello, this is a test response for compression!"
	if string(decompressed) != expected {
		t.Errorf("expected body %q, got %q", expected, string(decompressed))
	}
}

func TestCompressNoCompression(t *testing.T) {

	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler := Compress(http.HandlerFunc(mockHandler))

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected no Content-Encoding, got %q", rec.Header().Get("Content-Encoding"))
	}

	expected := "Hello, this is a test response for compression!"
	if rec.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, rec.Body.String())
	}
}

func TestCompressWebSocket(t *testing.T) {

	// Arrange
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	handler := Compress(http.HandlerFunc(mockHandler))

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected no Content-Encoding for websocket, got %q", rec.Header().Get("Content-Encoding"))
	}
}

func TestCompressContentTypeDetection(t *testing.T) {

	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	// no Content-Type handler
	handler := Compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("plain text content"))
	}))

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	contentType := rec.Header().Get("Content-Type")
	if contentType == "" {
		t.Error("expected Content-Type to be detected, got empty string")
	}
	// http.DetectContentType must detect text/plain
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("expected Content-Type to contain text/plain, got %q", contentType)
	}
}
