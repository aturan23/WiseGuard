package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// CalculateHash calculates the SHA-256 hash of the input
func CalculateHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// ValidateHashDifficulty checks if the hash has the required difficulty
func ValidateHashDifficulty(hash string, difficulty uint8) bool {
	prefix := strings.Repeat("0", int(difficulty))
	return strings.HasPrefix(hash, prefix)
}
