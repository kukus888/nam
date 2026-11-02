package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	data "kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func generateTestToken(t *testing.T, expiresIn time.Duration) string {
	t.Helper()

	claims := &services.Claims{
		Username: "alice",
		UserID:   42,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(services.GetJWTKeyProvider().Key)
	if err != nil {
		t.Fatalf("SignedString() returned error: %v", err)
	}

	return tokenString
}

func TestAuthMiddlewareRefreshesSoonExpiringTokenWithSecureCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		_, _ = c.Get("username")
		_, _ = c.Get("user_id")
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", generateTestToken(t, 2*time.Minute))
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("unexpected status code, got %d want %d", resp.Code, http.StatusNoContent)
	}

	setCookie := resp.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatal("expected Set-Cookie header to be present for refreshed token")
	}

	for _, expected := range []string{"token=", "Path=/", "HttpOnly", "Secure", "SameSite=Strict", "Max-Age=3600"} {
		if !strings.Contains(setCookie, expected) {
			t.Fatalf("expected Set-Cookie header to contain %q, got %q", expected, setCookie)
		}
	}
}

func TestAuthMiddlewareDoesNotRefreshTokenWhenNotNearExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", generateTestToken(t, 30*time.Minute))
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("unexpected status code, got %d want %d", resp.Code, http.StatusNoContent)
	}

	if setCookie := resp.Header().Get("Set-Cookie"); setCookie != "" {
		t.Fatalf("expected no Set-Cookie header for non-expiring token, got %q", setCookie)
	}
}

func TestGenerateTokenCompatibilityForMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	tokenString, err := services.GenerateToken(data.User{Id: 7, Username: "bob"})
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", tokenString)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("unexpected status code, got %d want %d", resp.Code, http.StatusNoContent)
	}
}
