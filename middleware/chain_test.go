package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewares(t *testing.T) {

	// Arrange
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "true")
			next.ServeHTTP(w, r)
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "true")
			next.ServeHTTP(w, r)
		})
	}

	m3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-3", "false")
			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Act
	wrapped := Middlewares(m1, m2, m3)(handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	// Assert
	if w.Header().Get("X-Middleware-1") != "true" {
		t.Errorf("Expected X-Middleware-1 header to be set")
	}

	if w.Header().Get("X-Middleware-2") != "true" {
		t.Errorf("Expected X-Middleware-2 header to be set")
	}

	if w.Header().Get("X-Middleware-3") != "false" {
		t.Errorf("Expected X-Middleware-3 header to be set")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestBind(t *testing.T) {

	// Arrange
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Bound", "true")
			next.ServeHTTP(w, r)
		})
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// Act
	boundHandler := Bind(middleware, handler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	boundHandler.ServeHTTP(w, req)

	// Assert
	if w.Header().Get("X-Bound") != "true" {
		t.Errorf("Expected X-Bound header to be set")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}
