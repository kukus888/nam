package services

import (
	"crypto/rand"
	"sync"
)

// JWTKeyProvider is a simple singleton to provide the JWT key for signing tokens.
type JWTKeyProvider struct {
	Key []byte
}

var lock = &sync.Once{}
var provider *JWTKeyProvider

// GetJWTKeyProvider returns the singleton instance of JWTKeyProvider.
// It generates a new key if it doesn't exist.
func GetJWTKeyProvider() *JWTKeyProvider {
	if provider == nil {
		lock.Do(func() {
			jwtKey := make([]byte, 384)
			rand.Read(jwtKey)
			provider = &JWTKeyProvider{
				Key: jwtKey,
			}
		})
	}
	return provider
}

// SetJWTKey allows setting a custom key for JWT signing.
// This is useful for testing or if you want to use a persistent key.
func SetJWTKey(key []byte) {
	lock.Do(func() {
		provider = &JWTKeyProvider{
			Key: key,
		}
	})
}
