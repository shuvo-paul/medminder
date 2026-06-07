package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequestEmailChange_Success(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	newEmail := "new@example.com"
	password := "Password123"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("RequestEmailChange", mock.Anything, userID, newEmail, password).Return(nil)

	handler := handlers.RequestEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.RequestEmailChangeInput{
		Authorization: "Bearer " + token,
	}
	input.Body.NewEmail = newEmail
	input.Body.CurrentPassword = password

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Verification email sent to your new email address", resp.Body.Message)
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestRequestEmailChange_InvalidAuthHeader(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	handler := handlers.RequestEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.RequestEmailChangeInput{
		Authorization: "Invalid",
	}
	input.Body.NewEmail = "new@example.com"
	input.Body.CurrentPassword = "Password123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired token")
}

func TestRequestEmailChange_EmptyNewEmail(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": uuid.New().String(),
	}, nil)

	handler := handlers.RequestEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.RequestEmailChangeInput{
		Authorization: "Bearer valid-token",
	}
	input.Body.NewEmail = ""
	input.Body.CurrentPassword = "Password123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "New email is required")
	mockTokenSvc.AssertExpectations(t)
}

func TestRequestEmailChange_EmptyPassword(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": uuid.New().String(),
	}, nil)

	handler := handlers.RequestEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.RequestEmailChangeInput{
		Authorization: "Bearer valid-token",
	}
	input.Body.NewEmail = "new@example.com"
	input.Body.CurrentPassword = ""

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Current password is required")
	mockTokenSvc.AssertExpectations(t)
}

func TestRequestEmailChange_EmailExists(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	newEmail := "new@example.com"
	password := "Password123"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("RequestEmailChange", mock.Anything, userID, newEmail, password).Return(service.ErrEmailExists)

	handler := handlers.RequestEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.RequestEmailChangeInput{
		Authorization: "Bearer " + token,
	}
	input.Body.NewEmail = newEmail
	input.Body.CurrentPassword = password

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Email already in use")
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestRequestEmailChange_WrongPassword(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	newEmail := "new@example.com"
	password := "Password123"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("RequestEmailChange", mock.Anything, userID, newEmail, password).Return(service.ErrWrongPassword)

	handler := handlers.RequestEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.RequestEmailChangeInput{
		Authorization: "Bearer " + token,
	}
	input.Body.NewEmail = newEmail
	input.Body.CurrentPassword = password

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Current password is incorrect")
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestRequestEmailChange_NoPasswordSet(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	newEmail := "new@example.com"
	password := "Password123"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("RequestEmailChange", mock.Anything, userID, newEmail, password).Return(service.ErrNoPasswordSet)

	handler := handlers.RequestEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.RequestEmailChangeInput{
		Authorization: "Bearer " + token,
	}
	input.Body.NewEmail = newEmail
	input.Body.CurrentPassword = password

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "no password set")
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestCancelEmailChange_Success(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("CancelEmailChange", mock.Anything, userID).Return(nil)

	handler := handlers.CancelEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.CancelEmailChangeInput{Authorization: "Bearer " + token}

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Email change request cancelled", resp.Body.Message)
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestCancelEmailChange_InvalidAuthHeader(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	handler := handlers.CancelEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.CancelEmailChangeInput{Authorization: "Invalid"}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired token")
}

func TestVerifyUpdatedEmail_Success(t *testing.T) {
	mockSvc := new(MockEmailChangeService)

	accessToken := "new-access-token"
	userID := uuid.New()
	newEmail := "new@example.com"

	mockSvc.On("VerifyEmailChange", mock.Anything, "valid-token").Return(&service.VerifyEmailChangeResult{
		AccessToken: accessToken,
		User: struct {
			ID            uuid.UUID
			Email         string
			DisplayName   string
			EmailVerified bool
		}{
			ID:            userID,
			Email:         newEmail,
			DisplayName:   "Test User",
			EmailVerified: true,
		},
	}, nil)

	handler := handlers.VerifyUpdatedEmailHandler(mockSvc)

	input := &dto.VerifyEmailInput{}
	input.Body.Token = "valid-token"

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, accessToken, resp.Body.AccessToken)
	mockSvc.AssertExpectations(t)
}

func TestVerifyUpdatedEmail_EmptyToken(t *testing.T) {
	mockSvc := new(MockEmailChangeService)

	handler := handlers.VerifyUpdatedEmailHandler(mockSvc)

	input := &dto.VerifyEmailInput{}
	input.Body.Token = ""

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Token is required")
}

func TestVerifyUpdatedEmail_InvalidToken(t *testing.T) {
	mockSvc := new(MockEmailChangeService)

	mockSvc.On("VerifyEmailChange", mock.Anything, "invalid-token").Return(nil, repository.ErrTokenNotFound)

	handler := handlers.VerifyUpdatedEmailHandler(mockSvc)

	input := &dto.VerifyEmailInput{}
	input.Body.Token = "invalid-token"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockSvc.AssertExpectations(t)
}

func TestVerifyUpdatedEmail_ExpiredToken(t *testing.T) {
	mockSvc := new(MockEmailChangeService)

	mockSvc.On("VerifyEmailChange", mock.Anything, "expired-token").Return(nil, repository.ErrTokenExpired)

	handler := handlers.VerifyUpdatedEmailHandler(mockSvc)

	input := &dto.VerifyEmailInput{}
	input.Body.Token = "expired-token"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockSvc.AssertExpectations(t)
}

func TestGetPendingEmailChange_Success(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	newEmail := "new@example.com"
	expiresAt := time.Now().Add(time.Hour)
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("GetPendingEmailChange", mock.Anything, userID).Return(newEmail, expiresAt, nil)

	handler := handlers.GetPendingEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.GetPendingEmailChangeInput{Authorization: "Bearer " + token}

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, newEmail, resp.Body.NewEmail)
	assert.Equal(t, expiresAt.Format(time.RFC3339), resp.Body.ExpiresAt)
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestGetPendingEmailChange_InvalidAuthHeader(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	handler := handlers.GetPendingEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.GetPendingEmailChangeInput{Authorization: "Invalid"}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired token")
}

func TestGetPendingEmailChange_NotFound(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("GetPendingEmailChange", mock.Anything, userID).Return("", time.Time{}, repository.ErrTokenNotFound)

	handler := handlers.GetPendingEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.GetPendingEmailChangeInput{Authorization: "Bearer " + token}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "No pending email change request")
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestGetPendingEmailChange_Expired(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("GetPendingEmailChange", mock.Anything, userID).Return("", time.Time{}, repository.ErrTokenExpired)

	handler := handlers.GetPendingEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.GetPendingEmailChangeInput{Authorization: "Bearer " + token}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "No pending email change request")
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}

func TestGetPendingEmailChange_ServiceError(t *testing.T) {
	mockSvc := new(MockEmailChangeService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.New()
	email := "test@example.com"
	token := "valid-access-token"

	mockTokenSvc.On("ValidateAccessToken", token).Return(jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
	}, nil)
	mockSvc.On("GetPendingEmailChange", mock.Anything, userID).Return("", time.Time{}, errors.New("database error"))

	handler := handlers.GetPendingEmailChangeHandler(mockSvc, mockTokenSvc)

	input := &dto.GetPendingEmailChangeInput{Authorization: "Bearer " + token}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockSvc.AssertExpectations(t)
	mockTokenSvc.AssertExpectations(t)
}
