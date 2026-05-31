package service_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-key"

func TestGenerateAccessToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"

	tokenSvc := service.NewTokenService(testJWTSecret)
	token, err := tokenSvc.GenerateAccessToken(userID, email)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := tokenSvc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims["sub"])
	assert.Equal(t, email, claims["email"])
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	tokenSvc := service.NewTokenService(testJWTSecret)

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"invalid", "invalid.token.here"},
		{"wrong signing key", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3OC0xMjM0LTEyMzQtMTIzNC0xMjM0NTY3ODkwYW4iLCJlbWFpbCI6ImV4YW1wbGVAZXhhbXBsZSJ9.invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tokenSvc.ValidateAccessToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestAccessTokenExpiry(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"

	tokenSvc := service.NewTokenService(testJWTSecret)
	token, err := tokenSvc.GenerateAccessToken(userID, email)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	claims, err := tokenSvc.ValidateAccessToken(token)
	require.NoError(t, err)

	exp, ok := claims["exp"].(float64)
	require.True(t, ok, "exp claim should be a number")

	expiry := time.Unix(int64(exp), 0)
	now := time.Now()

	assert.True(t, expiry.After(now), "token should not be expired")
	assert.WithinDuration(t, expiry, now.Add(24*time.Hour), 5*time.Second, "token should expire in ~24h")
}

func TestGenerateRefreshToken(t *testing.T) {
	tokenSvc := service.NewTokenService(testJWTSecret)
	token, err := tokenSvc.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 64, "refresh token should be 64 characters (32 bytes hex)")
}

func TestHashRefreshToken(t *testing.T) {
	tokenSvc := service.NewTokenService(testJWTSecret)
	token := "abc123"
	hash := tokenSvc.HashRefreshToken(token)
	assert.NotEqual(t, token, hash)
	assert.Len(t, hash, 64)
}

func TestVerifyRefreshToken(t *testing.T) {
	tokenSvc := service.NewTokenService(testJWTSecret)
	token := "abc123"
	hash := tokenSvc.HashRefreshToken(token)

	err := tokenSvc.VerifyRefreshToken(token, hash)
	assert.NoError(t, err)
}

func TestVerifyRefreshToken_Invalid(t *testing.T) {
	tokenSvc := service.NewTokenService(testJWTSecret)
	token := "abc123"
	hash := tokenSvc.HashRefreshToken(token)

	err := tokenSvc.VerifyRefreshToken("wrongtoken", hash)
	assert.Error(t, err)
}
