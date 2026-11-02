package services

import (
	"sync"
	"testing"
	"time"

	"kukus/nam/v2/layers/data"

	"github.com/golang-jwt/jwt/v5"
)

func setDeterministicJWTKey(t *testing.T) []byte {
	t.Helper()

	provider = nil
	lock = &sync.Once{}
	key := []byte("test-jwt-signing-key")
	SetJWTKey(key)
	return key
}

func TestGenerateTokenIncludesExpectedClaims(t *testing.T) {
	t.Parallel()

	key := setDeterministicJWTKey(t)
	user := data.User{Id: 42, Username: "alice"}

	tokenString, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	parsedClaims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, parsedClaims, func(token *jwt.Token) (any, error) {
		return key, nil
	})
	if err != nil {
		t.Fatalf("ParseWithClaims() returned error: %v", err)
	}
	if !token.Valid {
		t.Fatal("expected token to be valid")
	}

	if parsedClaims.Username != user.Username {
		t.Fatalf("username claim mismatch, got %q want %q", parsedClaims.Username, user.Username)
	}
	if parsedClaims.UserID != user.Id {
		t.Fatalf("user_id claim mismatch, got %d want %d", parsedClaims.UserID, user.Id)
	}

	if parsedClaims.ExpiresAt == nil {
		t.Fatal("expected expires_at claim to be set")
	}

	expiresIn := time.Until(parsedClaims.ExpiresAt.Time)
	if expiresIn < 55*time.Minute || expiresIn > 65*time.Minute {
		t.Fatalf("expected expiry around 1 hour, got %s", expiresIn)
	}
}

func TestRegenerateTokenExtendsExpiry(t *testing.T) {
	t.Parallel()

	key := setDeterministicJWTKey(t)
	oldExpiry := time.Now().Add(2 * time.Minute)
	claims := &Claims{
		Username: "alice",
		UserID:   7,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(oldExpiry),
		},
	}

	tokenString, err := RegenerateToken(claims)
	if err != nil {
		t.Fatalf("RegenerateToken() returned error: %v", err)
	}

	parsedClaims := &Claims{}
	_, err = jwt.ParseWithClaims(tokenString, parsedClaims, func(token *jwt.Token) (any, error) {
		return key, nil
	})
	if err != nil {
		t.Fatalf("ParseWithClaims() returned error: %v", err)
	}

	if !parsedClaims.ExpiresAt.Time.After(oldExpiry) {
		t.Fatalf("expected regenerated token expiry %s to be after old expiry %s", parsedClaims.ExpiresAt.Time, oldExpiry)
	}

	expiresIn := time.Until(parsedClaims.ExpiresAt.Time)
	if expiresIn < 55*time.Minute || expiresIn > 65*time.Minute {
		t.Fatalf("expected regenerated expiry around 1 hour, got %s", expiresIn)
	}
}
