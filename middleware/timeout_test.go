package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestContextTimeoutPassesRequestWhenHandlerFinishesBeforeDeadline(t *testing.T) {

	// Arrange
	handler := ContextTimeout(10 * time.Millisecond)(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", body)
	}
}

func TestContextTimeoutCancelsContext(t *testing.T) {

	// Arrange
	handler := ContextTimeout(10 * time.Millisecond)(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				w.WriteHeader(http.StatusRequestTimeout)
				_, _ = w.Write([]byte("timeout"))

			case <-time.After(20 * time.Millisecond):
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("late"))
			}
		},
	))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusRequestTimeout, rec.Code)
	}

	if body := rec.Body.String(); body != "timeout" {
		t.Fatalf("expected body %q, got %q", "timeout", body)
	}
}

func TestContextTimeoutRespectsExistingDeadline(t *testing.T) {

	// Arrange
	var gotDeadline time.Time
	var hasDeadline bool

	handler := ContextTimeout(30 * time.Millisecond)(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gotDeadline, hasDeadline = r.Context().Deadline()
			w.WriteHeader(http.StatusOK)
		},
	))

	baseDeadline := time.Now().Add(10 * time.Millisecond)
	baseCtx, cancel := context.WithDeadline(context.Background(), baseDeadline)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/deadline", nil).WithContext(baseCtx)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if !hasDeadline {
		t.Fatal("expected request context to have deadline")
	}

	if gotDeadline.After(baseDeadline.Add(time.Millisecond)) {
		t.Fatalf("expected deadline not after base deadline %v, got %v", baseDeadline, gotDeadline)
	}
}
