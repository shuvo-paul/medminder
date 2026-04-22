package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Successful(t *testing.T) {
	mockRepo := new(MockRegisterUserRepository)

	userID := uuid.New()
	email := "test@example.com"
	password := "Password123"
	displayName := "Test User"

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(db.User{}, sql.ErrNoRows)
	mockRepo.On("CreateUser", mock.Anything, email, displayName, mock.Anything).Return(db.CreateUserRow{
		ID:            userID,
		Email:         email,
		DisplayName:   displayName,
		EmailVerified: sql.NullBool{Bool: false, Valid: true},
	}, nil)

	handler := handlers.RegisterHandler(mockRepo)

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
	mockRepo := new(MockRegisterUserRepository)
	handler := handlers.RegisterHandler(mockRepo)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       "invalid-email",
		DisplayName: "Test User",
		Password:    "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRegister_InvalidPassword(t *testing.T) {
	mockRepo := new(MockRegisterUserRepository)
	handler := handlers.RegisterHandler(mockRepo)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
		Password:    "weak",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockRegisterUserRepository)

	email := "test@example.com"
	userID := uuid.New()

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(db.User{
		ID:    userID,
		Email: email,
	}, nil)

	handler := handlers.RegisterHandler(mockRepo)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       email,
		DisplayName: "Test User",
		Password:    "Password123",
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, handlers.ErrEmailExists))
	assert.Nil(t, resp)
}
