package services

import (
	"crypto/rand"
	"sync"
)

// JWTKeyProvider is a simple singleton to provide the JWT key for signing tokens.
type JWTKeyProvider struct {
	Key []byte
}

var jwtLock = &sync.Once{}
var jwtProvider *JWTKeyProvider

// GetJWTKeyProvider returns the singleton instance of JWTKeyProvider.
// It generates a new key if it doesn't exist.
func GetJWTKeyProvider() *JWTKeyProvider {
	if jwtProvider == nil {
		jwtLock.Do(func() {
			jwtKey := make([]byte, 384)
			rand.Read(jwtKey)
			jwtProvider = &JWTKeyProvider{
				Key: jwtKey,
			}
		})
	}
	return jwtProvider
}

// SetJWTKey allows setting a custom key for JWT signing.
// This is useful for testing or if you want to use a persistent key.
func SetJWTKey(key []byte) {
	jwtLock.Do(func() {
		jwtProvider = &JWTKeyProvider{
			Key: key,
		}
	})
}
