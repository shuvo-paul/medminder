package handlers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Successful(t *testing.T) {
	mockSvc := new(MockAuthService)

	userID := uuid.New()
	email := "test@example.com"
	password := "Password123"
	displayName := "Test User"

	mockSvc.On("Register", mock.Anything, email, displayName, password).Return(&service.RegisterResult{
		User: struct {
			ID            uuid.UUID
			Email         string
			DisplayName   string
			EmailVerified bool
		}{
			ID:            userID,
			Email:         email,
			DisplayName:   displayName,
			EmailVerified: false,
		},
	}, nil)

	handler := handlers.RegisterHandler(mockSvc)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       email,
		DisplayName: displayName,
		Password:    password,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.User.ID)
	assert.Equal(t, userID, resp.User.ID)
	assert.Equal(t, email, resp.User.Email)
	assert.Equal(t, displayName, resp.User.DisplayName)
	assert.False(t, resp.User.EmailVerified)
}

func TestRegister_InvalidEmail(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := handlers.RegisterHandler(mockSvc)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       "invalid-email",
		DisplayName: "Test User",
		Password:    "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRegister_InvalidPassword(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := handlers.RegisterHandler(mockSvc)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
		Password:    "weak",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}