package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
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

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    email,
		Password: password,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, userID, resp.User.ID)
	assert.Equal(t, email, resp.User.Email)
	assert.Equal(t, displayName, resp.User.DisplayName)
	assert.True(t, resp.User.EmailVerified)
}

func TestLogin_InvalidEmail(t *testing.T) {
	mockSvc := new(MockAuthService)

	handler := handlers.LoginHandler(mockSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "invalid-email",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidEmail))
}

func TestLogin_UserNotFound(t *testing.T) {
	mockSvc := new(MockAuthService)

	mockSvc.On("Login", mock.Anything, "nonexistent@example.com", "Password123").Return(nil, service.ErrInvalidCredentials)

	handler := handlers.LoginHandler(mockSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestLogin_EmptyPasswordHash(t *testing.T) {
	mockSvc := new(MockAuthService)

	mockSvc.On("Login", mock.Anything, "test@example.com", "Password123").Return(nil, service.ErrInvalidCredentials)

	handler := handlers.LoginHandler(mockSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "test@example.com",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockSvc := new(MockAuthService)

	mockSvc.On("Login", mock.Anything, "test@example.com", "WrongPassword123").Return(nil, service.ErrInvalidCredentials)

	handler := handlers.LoginHandler(mockSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "test@example.com",
		Password: "WrongPassword123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}