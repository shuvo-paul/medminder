package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogout_Successful(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockAuthSvc.On("Logout", mock.Anything, userID).Return(nil)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	handler := handlers.LogoutHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.LogoutInput{Authorization: "Bearer " + token}
	_, err := handler(context.Background(), input)

	assert.NoError(t, err)
	mockAuthSvc.AssertExpectations(t)
}

func TestLogout_InvalidHeader(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockTokenSvc := new(MockTokenService)

	handler := handlers.LogoutHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.LogoutInput{Authorization: "Invalid"}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid authorization header")
}

func TestLogout_ExpiredToken(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(-time.Hour))

	mockTokenSvc.On("ValidateAccessToken", token).Return(nil, service.ErrTokenExpired)

	handler := handlers.LogoutHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.LogoutInput{Authorization: "Bearer " + token}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid or expired access token")
}

func TestLogout_DBError(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockAuthSvc.On("Logout", mock.Anything, userID).Return(service.ErrLogoutFailed)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	handler := handlers.LogoutHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.LogoutInput{Authorization: "Bearer " + token}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
}

func makeAccessToken(userID uuid.UUID, email string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"exp":   exp.Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}
