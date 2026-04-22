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
	"golang.org/x/crypto/bcrypt"
)

func TestLogin_Successful(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockTokenSvc := new(MockLoginTokenService)

	userID := uuid.New()
	email := "test@example.com"
	password := "Password123"
	displayName := "Test User"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	mockUserRepo.On("GetUserByEmail", mock.Anything, email).Return(db.User{
		ID:            userID,
		Email:         email,
		DisplayName:   displayName,
		PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}, nil)
	mockTokenSvc.On("GenerateAccessToken", userID, email).Return("access-token", nil)
	mockTokenSvc.On("GenerateRefreshToken").Return("refresh-token", nil)
	mockTokenSvc.On("HashRefreshToken", "refresh-token").Return("hashed-token")
	mockTokenRepo.On("CreateRefreshToken", mock.Anything, userID, "hashed-token", mock.Anything).Return(db.CreateRefreshTokenRow{}, nil)

	handler := handlers.LoginHandler(mockUserRepo, mockTokenRepo, mockTokenSvc)

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
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockTokenSvc := new(MockLoginTokenService)
	handler := handlers.LoginHandler(mockUserRepo, mockTokenRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "invalid-email",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidEmail))
}

func TestLogin_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockTokenSvc := new(MockLoginTokenService)

	mockUserRepo.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return(db.User{}, sql.ErrNoRows)

	handler := handlers.LoginHandler(mockUserRepo, mockTokenRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidCredentials))
}

func TestLogin_EmptyPasswordHash(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockTokenSvc := new(MockLoginTokenService)

	userID := uuid.New()

	mockUserRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(db.User{
		ID:            userID,
		Email:         "test@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{Valid: false},
		EmailVerified: sql.NullBool{Bool: false, Valid: true},
	}, nil)

	handler := handlers.LoginHandler(mockUserRepo, mockTokenRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "test@example.com",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidCredentials))
}

func TestLogin_WrongPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockTokenSvc := new(MockLoginTokenService)

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("CorrectPassword123"), 12)

	mockUserRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(db.User{
		ID:            userID,
		Email:         "test@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
		EmailVerified: sql.NullBool{Bool: false, Valid: true},
	}, nil)

	handler := handlers.LoginHandler(mockUserRepo, mockTokenRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "test@example.com",
		Password: "WrongPassword123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidCredentials))
}
