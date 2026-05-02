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

	input := &dto.RegisterInput{}
	input.Body.Email = email
	input.Body.DisplayName = displayName
	input.Body.Password = password

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Body.User.ID)
	assert.Equal(t, userID, resp.Body.User.ID)
	assert.Equal(t, email, resp.Body.User.Email)
	assert.Equal(t, displayName, resp.Body.User.DisplayName)
	assert.False(t, resp.Body.User.EmailVerified)
}

func TestRegister_InvalidEmail(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := handlers.RegisterHandler(mockSvc)

	input := &dto.RegisterInput{}
	input.Body.Email = "invalid-email"
	input.Body.DisplayName = "Test User"
	input.Body.Password = "Password123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid email")
}

func TestRegister_InvalidPassword(t *testing.T) {
	mockSvc := new(MockAuthService)
	handler := handlers.RegisterHandler(mockSvc)

	input := &dto.RegisterInput{}
	input.Body.Email = "test@example.com"
	input.Body.DisplayName = "Test User"
	input.Body.Password = "weak"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Password must be at least")
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mockSvc := new(MockAuthService)

	mockSvc.On("Register", mock.Anything, "test@example.com", "Test User", "Password123").
		Return(nil, service.ErrEmailExists)

	handler := handlers.RegisterHandler(mockSvc)

	input := &dto.RegisterInput{}
	input.Body.Email = "test@example.com"
	input.Body.DisplayName = "Test User"
	input.Body.Password = "Password123"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Email already exists")
}
