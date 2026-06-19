package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestRequestIDGeneratesNewID(t *testing.T) {

	// Arrange
	handlerCalled := false
	var capturedID string

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// Get ID from context
		id := IDFromContext(r.Context())
		if id == "" {
			t.Error("Expected request ID to be set in context, got empty string")
		}
		capturedID = id

		// Check if the captured ID is a valid UUID
		_, err := uuid.Parse(id)
		if err != nil {
			t.Errorf("Expected valid UUID, got %q: %v", id, err)
		}
	})

	// Act
	req := httptest.NewRequestWithContext(context.TODO(), "GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler := RequestID(testHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	// Check if X-Request-ID header is set in response
	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set in response")
	}

	// Check if the captured ID matches the expected one
	if requestID != capturedID {
		t.Errorf("Expected X-Request-ID header %q to match context ID %q", requestID, capturedID)
	}
}

func TestRequestIDUsesExistingID(t *testing.T) {

	// Arrange
	existingID := "test-request-123"
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// Set ID in context
		id := IDFromContext(r.Context())
		if id != existingID {
			t.Errorf("Expected request ID %q, got %q", existingID, id)
		}
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	rec := httptest.NewRecorder()
	handler := RequestID(testHandler)
	handler.ServeHTTP(rec, req)

	// Assert
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	// Check response header
	responseID := rec.Header().Get("X-Request-ID")
	if responseID != existingID {
		t.Errorf("Expected X-Request-ID header %q, got %q", existingID, responseID)
	}
}

func TestRequestIDPreservesContext(t *testing.T) {

	// Arrange
	type ctxKey string
	const customKey ctxKey = "custom_key"
	customValue := "custom_value"

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// check context for custom key
		val := r.Context().Value(customKey)
		if val != customValue {
			t.Errorf("Expected custom context value %q, got %q", customValue, val)
		}

		// Set request ID in the context
		id := IDFromContext(r.Context())
		if id == "" {
			t.Error("Expected request ID to be set in context")
		}
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)

	// Set custom key in the context
	ctx := context.WithValue(req.Context(), customKey, customValue)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler := RequestID(testHandler)
	handler.ServeHTTP(rec, req)
}

func TestIDFromContextReturnsEmptyForMissingKey(t *testing.T) {

	// Arrange
	ctx := context.Background()

	// Act
	id := IDFromContext(ctx)

	// Assert
	if id != "" {
		t.Errorf("Expected empty string for missing key, got %q", id)
	}
}

func TestIDFromContextReturnsEmptyForWrongType(t *testing.T) {

	// Arrange
	ctx := context.WithValue(context.Background(), RequestIDKey, 12345)

	// Act
	id := IDFromContext(ctx)

	// Assert
	if id != "" {
		t.Errorf("Expected empty string for wrong type, got %q", id)
	}
}
