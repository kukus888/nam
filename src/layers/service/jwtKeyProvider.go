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
