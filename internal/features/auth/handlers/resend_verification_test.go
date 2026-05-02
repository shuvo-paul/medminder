package handlers_test

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResendVerification_Successful(t *testing.T) {
	mockSvc := new(MockEmailVerificationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":  userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("ResendVerification", mock.Anything, userID, email).Return(nil)

	handler := handlers.ResendVerificationHandler(mockSvc, mockTokenSvc)

	input := &dto.ResendVerificationInput{}
	input.Body.Authorization = "Bearer " + token

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Verification email sent", resp.Body.Message)
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestResendVerification_InvalidAuthHeader(t *testing.T) {
	mockSvc := new(MockEmailVerificationService)
	mockTokenSvc := new(MockTokenService)
	handler := handlers.ResendVerificationHandler(mockSvc, mockTokenSvc)

	input := &dto.ResendVerificationInput{}
	input.Body.Authorization = "Invalid"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid authorization header")
}

func TestResendVerification_InvalidToken(t *testing.T) {
	mockSvc := new(MockEmailVerificationService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "invalid-token").Return(nil, service.ErrInvalidToken)

	handler := handlers.ResendVerificationHandler(mockSvc, mockTokenSvc)

	input := &dto.ResendVerificationInput{}
	input.Body.Authorization = "Bearer invalid-token"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockTokenSvc.AssertExpectations(t)
}

func TestResendVerification_RateLimitExceeded(t *testing.T) {
	mockSvc := new(MockEmailVerificationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("ResendVerification", mock.Anything, userID, email).Return(service.ErrRateLimitExceeded)

	handler := handlers.ResendVerificationHandler(mockSvc, mockTokenSvc)

	input := &dto.ResendVerificationInput{}
	input.Body.Authorization = "Bearer " + token

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Daily verification limit exceeded")
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}