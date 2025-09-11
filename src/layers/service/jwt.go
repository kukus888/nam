package services

import (
	"kukus/nam/v2/layers/data"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	UserID   uint64 `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(user data.User) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		UserID:   user.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetJWTKeyProvider().Key)
}
