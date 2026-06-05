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
	"github.com/stretchr/testify/require"
)

func TestSetPasswordHandler_Success(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)
	mockAuthSvc.On("SetPassword", mock.Anything, userID, "NewPwd123").Return(nil)

	handler := handlers.SetPasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.SetPasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.Password = "NewPwd123"

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "password set successfully", resp.Body.Message)
	mockAuthSvc.AssertExpectations(t)
}

func TestSetPasswordHandler_InvalidToken(t *testing.T) {
	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", "bad-token").Return(nil, service.ErrInvalidToken)

	mockAuthSvc := new(MockAuthService)

	handler := handlers.SetPasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.SetPasswordInput{
		Authorization: "Bearer bad-token",
	}
	input.Body.Password = "NewPwd123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired")
}

func TestSetPasswordHandler_WeakPassword(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)

	handler := handlers.SetPasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.SetPasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.Password = "weak"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid password")
	mockAuthSvc.AssertNotCalled(t, "SetPassword", mock.Anything, mock.Anything, mock.Anything)
}

func TestSetPasswordHandler_UpdateFailed(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)
	mockAuthSvc.On("SetPassword", mock.Anything, userID, "NewPwd123").Return(assert.AnError)

	handler := handlers.SetPasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.SetPasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.Password = "NewPwd123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to set password")
}

func TestChangePasswordHandler_Success(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)
	mockAuthSvc.On("ChangePassword", mock.Anything, userID, "CurrentPwd1", "NewPwd123").Return(nil)

	handler := handlers.ChangePasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.ChangePasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.CurrentPassword = "CurrentPwd1"
	input.Body.NewPassword = "NewPwd123"
	input.Body.ConfirmPassword = "NewPwd123"

	resp, err := handler(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "password changed successfully", resp.Body.Message)
	mockAuthSvc.AssertExpectations(t)
}

func TestChangePasswordHandler_InvalidToken(t *testing.T) {
	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", "bad-token").Return(nil, service.ErrInvalidToken)

	mockAuthSvc := new(MockAuthService)

	handler := handlers.ChangePasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.ChangePasswordInput{
		Authorization: "Bearer bad-token",
	}
	input.Body.CurrentPassword = "CurrentPwd1"
	input.Body.NewPassword = "NewPwd123"
	input.Body.ConfirmPassword = "NewPwd123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired")
}

func TestChangePasswordHandler_WeakNewPassword(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)

	handler := handlers.ChangePasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.ChangePasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.CurrentPassword = "CurrentPwd1"
	input.Body.NewPassword = "short"
	input.Body.ConfirmPassword = "short"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid new password")
	mockAuthSvc.AssertNotCalled(t, "ChangePassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestChangePasswordHandler_PasswordMismatch(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)

	handler := handlers.ChangePasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.ChangePasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.CurrentPassword = "CurrentPwd1"
	input.Body.NewPassword = "NewPwd123"
	input.Body.ConfirmPassword = "DifferentPwd1"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "passwords do not match")
	mockAuthSvc.AssertNotCalled(t, "ChangePassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestChangePasswordHandler_WrongCurrentPassword(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)
	mockAuthSvc.On("ChangePassword", mock.Anything, userID, "WrongPwd1", "NewPwd123").
		Return(service.ErrWrongPassword)

	handler := handlers.ChangePasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.ChangePasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.CurrentPassword = "WrongPwd1"
	input.Body.NewPassword = "NewPwd123"
	input.Body.ConfirmPassword = "NewPwd123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "current password is incorrect")
}

func TestChangePasswordHandler_NoPasswordSet(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)
	mockAuthSvc.On("ChangePassword", mock.Anything, userID, "CurrentPwd1", "NewPwd123").
		Return(service.ErrNoPasswordSet)

	handler := handlers.ChangePasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.ChangePasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.CurrentPassword = "CurrentPwd1"
	input.Body.NewPassword = "NewPwd123"
	input.Body.ConfirmPassword = "NewPwd123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "no password set")
}

func TestChangePasswordHandler_ServiceError(t *testing.T) {
	userID := uuid.New()
	token := makeAccessToken(userID, "test@example.com", time.Now().Add(time.Hour))

	mockTokenSvc := new(MockTokenService)
	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{"sub": userID.String()}, nil)

	mockAuthSvc := new(MockAuthService)
	mockAuthSvc.On("ChangePassword", mock.Anything, userID, "CurrentPwd1", "NewPwd123").
		Return(assert.AnError)

	handler := handlers.ChangePasswordHandler(mockAuthSvc, mockTokenSvc)

	input := &dto.ChangePasswordInput{
		Authorization: "Bearer " + token,
	}
	input.Body.CurrentPassword = "CurrentPwd1"
	input.Body.NewPassword = "NewPwd123"
	input.Body.ConfirmPassword = "NewPwd123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to change password")
}
