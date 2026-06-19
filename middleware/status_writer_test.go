package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteHeaderSingleCall(t *testing.T) {

	// Arrange
	rr := httptest.NewRecorder()
	sr := NewStatusHTTP(rr)

	// Act
	sr.WriteHeader(http.StatusNotFound)

	// Assert
	if sr.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, sr.StatusCode)
	}

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected response recorder code %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestWriteHeaderMultipleCalls(t *testing.T) {

	// Arrange
	rr := httptest.NewRecorder()
	sr := NewStatusHTTP(rr)

	// Act
	sr.WriteHeader(http.StatusNotFound)
	sr.WriteHeader(http.StatusInternalServerError) // must be ignored

	// Assert
	if sr.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, sr.StatusCode)
	}

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected response recorder code %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestDefaultStatusCode(t *testing.T) {

	// Arrange
	rr := httptest.NewRecorder()
	sr := NewStatusHTTP(rr)

	// Act
	// don't call WriteHeader, just call Write
	sr.Write([]byte("OK"))

	// Assert
	if sr.StatusCode != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, sr.StatusCode)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected response recorder code %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestWriteHeaderValid(t *testing.T) {

	// Arrange
	sr := StatusHTTP{
		httptest.NewRecorder(),
		http.StatusOK,
		false,
	}
	testStatusCode := 400

	// Act
	sr.WriteHeader(testStatusCode)

	// Assert
	if sr.StatusCode != testStatusCode {
		t.Errorf("Wrong Status Code in header!: got %d", sr.StatusCode)
	}
}

func TestNotEmptyNewStatusHTTP(t *testing.T) {

	// Arrange
	rr := httptest.NewRecorder()

	// Act
	result := NewStatusHTTP(rr)

	// Assert
	if result == nil {
		t.Errorf("NewStatusHTTP must return non nil!: got %v", result)
	}
}
