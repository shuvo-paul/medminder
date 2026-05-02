package handlers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogin_Successful(t *testing.T) {
	mockSvc := new(MockAuthService)

	userID := uuid.New()
	email := "test@example.com"
	password := "Password123"
	displayName := "Test User"

	mockSvc.On("Login", mock.Anything, email, password).Return(&service.LoginResult{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		User: struct {
			ID            uuid.UUID
			Email         string
			DisplayName   string
			EmailVerified bool
		}{
			ID:            userID,
			Email:         email,
			DisplayName:   displayName,
			EmailVerified: true,
		},
	}, nil)

	handler := handlers.LoginHandler(mockSvc)

	input := &dto.LoginInput{}
	input.Body.Email = email
	input.Body.Password = password

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Body.AccessToken)
	assert.NotEmpty(t, resp.Body.RefreshToken)
	assert.Equal(t, userID, resp.Body.User.ID)
	assert.Equal(t, email, resp.Body.User.Email)
	assert.Equal(t, displayName, resp.Body.User.DisplayName)
	assert.True(t, resp.Body.User.EmailVerified)
}

func TestLogin_InvalidEmail(t *testing.T) {
	mockSvc := new(MockAuthService)

	handler := handlers.LoginHandler(mockSvc)

	input := &dto.LoginInput{}
	input.Body.Email = "invalid-email"
	input.Body.Password = "Password123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid email format")
}

func TestLogin_UserNotFound(t *testing.T) {
	mockSvc := new(MockAuthService)

	mockSvc.On("Login", mock.Anything, "nonexistent@example.com", "Password123").Return(nil, service.ErrInvalidCredentials)

	handler := handlers.LoginHandler(mockSvc)

	input := &dto.LoginInput{}
	input.Body.Email = "nonexistent@example.com"
	input.Body.Password = "Password123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestLogin_EmptyPasswordHash(t *testing.T) {
	mockSvc := new(MockAuthService)

	mockSvc.On("Login", mock.Anything, "test@example.com", "Password123").Return(nil, service.ErrInvalidCredentials)

	handler := handlers.LoginHandler(mockSvc)

	input := &dto.LoginInput{}
	input.Body.Email = "test@example.com"
	input.Body.Password = "Password123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockSvc := new(MockAuthService)

	mockSvc.On("Login", mock.Anything, "test@example.com", "WrongPassword123").Return(nil, service.ErrInvalidCredentials)

	handler := handlers.LoginHandler(mockSvc)

	input := &dto.LoginInput{}
	input.Body.Email = "test@example.com"
	input.Body.Password = "WrongPassword123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}
