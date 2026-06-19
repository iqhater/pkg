package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverOK(t *testing.T) {

	// Arrange
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test recover middleware panic")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	// Act
	Recover(next).ServeHTTP(w, r)

	// Assert
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Recover() = %v, want %v", w.Code, http.StatusInternalServerError)
	}
}
