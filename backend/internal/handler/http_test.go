package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auction-platform/internal/config"
	"auction-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthRequiredRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	h := NewHandler(newTestAuthService(), nil, nil, nil, nil)
	router.GET("/protected", h.AuthRequired(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthRequiredAcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	h := NewHandler(newTestAuthService(), nil, nil, nil, nil)
	router.GET("/protected", h.AuthRequired(), func(c *gin.Context) {
		if got := h.getUserID(c); got != 42 {
			t.Fatalf("expected user id 42, got %d", got)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedTestToken(t, 42))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func newTestAuthService() *service.AuthService {
	return service.NewAuthService(nil, &config.JWTConfig{
		Secret:          "test-secret",
		ExpirationHours: 1,
	})
}

func signedTestToken(t *testing.T, userID int64) string {
	t.Helper()

	claims := &service.Claims{
		UserID:   userID,
		Username: "tester",
		Role:     "buyer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}
