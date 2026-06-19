package generate

import (
	"testing"
)

func TestGenerateRandomTokenLength(t *testing.T) {

	// Arrange
	length := uint(16)

	// Act
	token, _ := GenerateRandomToken(length)

	// Assert
	// Check that the token length is correct (each byte is represented by 2 hex characters)
	expectedLength := int(length * 2)
	if len(token) != expectedLength {
		t.Fatalf("Token length should be twice the input length (hex representation): got %d, want %d", len(token), expectedLength)
	}
}

func TestGenerateRandomTokenUnique(t *testing.T) {

	// Arrange
	length := uint(16)

	// Act
	token1, _ := GenerateRandomToken(length)
	token2, _ := GenerateRandomToken(length)

	// Assert
	// Check that the tokens are unique
	if token1 == token2 {
		t.Fatal("Two tokens should not be the same")
	}
	if token1 == "" {
		t.Fatal("Generated token 1 should not be empty")
	}
	if token2 == "" {
		t.Fatal("Generated token 2 should not be empty")
	}
}

func TestGenerateRandomTokenZeroLength(t *testing.T) {

	// Arrange
	length := uint(0)

	// Act
	token, _ := GenerateRandomToken(length)

	// Assert
	if token != "" {
		t.Fatalf("Token should be an empty string when length is 0: got %q", token)
	}
}

func TestGenerateRandomTokenHexFormat(t *testing.T) {

	// Arrange
	length := uint(16)

	// Act
	token, _ := GenerateRandomToken(length)

	// Assert
	// Check that the token only contains hexadecimal characters
	for _, r := range token {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Fatalf("Token should only contain hexadecimal characters: got %q", r)
		}
	}
}
