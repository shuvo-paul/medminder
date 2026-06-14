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

func TestDeleteAccount_Successful(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockProfileSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)
	mockAuthSvc.On("VerifyPassword", mock.Anything, userID, "Password123").Return(nil)
	mockProfileSvc.On("HandleAccountDeletion", mock.Anything, userID).Return(nil)
	mockAuthSvc.On("DeleteAccount", mock.Anything, userID).Return(nil)

	handler := handlers.DeleteAccountHandler(mockAuthSvc, mockProfileSvc, mockTokenSvc)

	input := &dto.DeleteAccountInput{
		Authorization: "Bearer " + token,
		Body: struct {
			Password string `json:"password" minLength:"1"`
		}{
			Password: "Password123",
		},
	}
	_, err := handler(context.Background(), input)

	assert.NoError(t, err)
	mockAuthSvc.AssertExpectations(t)
	mockProfileSvc.AssertExpectations(t)
}

func TestDeleteAccount_InvalidHeader(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockProfileSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	handler := handlers.DeleteAccountHandler(mockAuthSvc, mockProfileSvc, mockTokenSvc)

	input := &dto.DeleteAccountInput{
		Authorization: "Invalid",
		Body: struct {
			Password string `json:"password" minLength:"1"`
		}{
			Password: "Password123",
		},
	}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid authorization header")
}

func TestDeleteAccount_WrongPassword(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockProfileSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)
	mockAuthSvc.On("VerifyPassword", mock.Anything, userID, "wrong-password").Return(service.ErrWrongPassword)

	handler := handlers.DeleteAccountHandler(mockAuthSvc, mockProfileSvc, mockTokenSvc)

	input := &dto.DeleteAccountInput{
		Authorization: "Bearer " + token,
		Body: struct {
			Password string `json:"password" minLength:"1"`
		}{
			Password: "wrong-password",
		},
	}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "current password is incorrect")
}

func TestDeleteAccount_UserNotFound(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockProfileSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)
	mockAuthSvc.On("VerifyPassword", mock.Anything, userID, "Password123").Return(service.ErrUserNotFound)

	handler := handlers.DeleteAccountHandler(mockAuthSvc, mockProfileSvc, mockTokenSvc)

	input := &dto.DeleteAccountInput{
		Authorization: "Bearer " + token,
		Body: struct {
			Password string `json:"password" minLength:"1"`
		}{
			Password: "Password123",
		},
	}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestDeleteAccount_ProfileServiceError(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockProfileSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)
	mockAuthSvc.On("VerifyPassword", mock.Anything, userID, "Password123").Return(nil)
	mockProfileSvc.On("HandleAccountDeletion", mock.Anything, userID).Return(assert.AnError)

	handler := handlers.DeleteAccountHandler(mockAuthSvc, mockProfileSvc, mockTokenSvc)

	input := &dto.DeleteAccountInput{
		Authorization: "Bearer " + token,
		Body: struct {
			Password string `json:"password" minLength:"1"`
		}{
			Password: "Password123",
		},
	}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to handle profile cleanup")
}

func TestDeleteAccount_DeleteAccountError(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockProfileSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)
	mockAuthSvc.On("VerifyPassword", mock.Anything, userID, "Password123").Return(nil)
	mockProfileSvc.On("HandleAccountDeletion", mock.Anything, userID).Return(nil)
	mockAuthSvc.On("DeleteAccount", mock.Anything, userID).Return(assert.AnError)

	handler := handlers.DeleteAccountHandler(mockAuthSvc, mockProfileSvc, mockTokenSvc)

	input := &dto.DeleteAccountInput{
		Authorization: "Bearer " + token,
		Body: struct {
			Password string `json:"password" minLength:"1"`
		}{
			Password: "Password123",
		},
	}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete account")
}

func TestDeleteAccount_ExpiredToken(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockProfileSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(-time.Hour))

	mockTokenSvc.On("ValidateAccessToken", token).Return(nil, service.ErrTokenExpired)

	handler := handlers.DeleteAccountHandler(mockAuthSvc, mockProfileSvc, mockTokenSvc)

	input := &dto.DeleteAccountInput{
		Authorization: "Bearer " + token,
		Body: struct {
			Password string `json:"password" minLength:"1"`
		}{
			Password: "Password123",
		},
	}
	_, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid or expired access token")
}
