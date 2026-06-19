package generate

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken method generate random string token
func GenerateRandomToken(length uint) (string, error) {

	b := make([]byte, length)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b), nil
}
