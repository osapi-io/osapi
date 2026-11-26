package seal

import (
	"crypto/ecdh"
	"crypto/rand"
)

// SetCreateKey replaces the function that makes a P-256 key.
func SetCreateKey(
	fn func() (*ecdh.PrivateKey, error),
) {
	createKey = fn
}

// ResetCreateKey restores the function that makes a P-256 key.
func ResetCreateKey() {
	createKey = func() (*ecdh.PrivateKey, error) {
		return ecdh.P256().GenerateKey(rand.Reader)
	}
}
