package utils

import "crypto/subtle"

// SecureCompareHash performs constant-time comparison of two hash strings
// to prevent timing attacks when comparing tokens or API keys
func SecureCompareHash(hash1, hash2 string) bool {
	// Convert strings to byte slices
	b1 := []byte(hash1)
	b2 := []byte(hash2)

	// subtle.ConstantTimeCompare returns 1 if equal, 0 if not
	// It takes constant time regardless of where the difference is
	return subtle.ConstantTimeCompare(b1, b2) == 1
}
