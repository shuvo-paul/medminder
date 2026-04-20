package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockLoginUserRepository struct {
	mock.Mock
}

func (m *MockLoginUserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockLoginUserRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error) {
	args := m.Called(ctx, userID, tokenHash, expiresAt)
	return args.Get(0).(db.CreateRefreshTokenRow), args.Error(1)
}

type MockLoginTokenService struct {
	mock.Mock
}

func (m *MockLoginTokenService) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockLoginTokenService) GenerateRefreshToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockLoginTokenService) HashRefreshToken(token string) string {
	args := m.Called(token)
	return args.String(0)
}

func TestLogin_Successful(t *testing.T) {
	mockRepo := new(MockLoginUserRepository)
	mockTokenSvc := new(MockLoginTokenService)

	userID := uuid.New()
	email := "test@example.com"
	password := "Password123"
	displayName := "Test User"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(db.User{
		ID:            userID,
		Email:         email,
		DisplayName:   displayName,
		PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}, nil)
	mockTokenSvc.On("GenerateAccessToken", userID, email).Return("access-token", nil)
	mockTokenSvc.On("GenerateRefreshToken").Return("refresh-token", nil)
	mockTokenSvc.On("HashRefreshToken", "refresh-token").Return("hashed-token")
	mockRepo.On("CreateRefreshToken", mock.Anything, userID, "hashed-token", mock.Anything).Return(db.CreateRefreshTokenRow{}, nil)

	handler := handlers.LoginHandler(mockRepo, mockTokenSvc)

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
	mockRepo := new(MockLoginUserRepository)
	mockTokenSvc := new(MockLoginTokenService)
	handler := handlers.LoginHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "invalid-email",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidEmail))
}

func TestLogin_EmptyPassword(t *testing.T) {
	mockRepo := new(MockLoginUserRepository)
	mockTokenSvc := new(MockLoginTokenService)
	handler := handlers.LoginHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "test@example.com",
		Password: "",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidCredentials))
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(MockLoginUserRepository)
	mockTokenSvc := new(MockLoginTokenService)

	mockRepo.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return(db.User{}, sql.ErrNoRows)

	handler := handlers.LoginHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidCredentials))
}

func TestLogin_EmptyPasswordHash(t *testing.T) {
	mockRepo := new(MockLoginUserRepository)
	mockTokenSvc := new(MockLoginTokenService)

	userID := uuid.New()

	mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(db.User{
		ID:            userID,
		Email:         "test@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{Valid: false},
		EmailVerified: sql.NullBool{Bool: false, Valid: true},
	}, nil)

	handler := handlers.LoginHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "test@example.com",
		Password: "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidCredentials))
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(MockLoginUserRepository)
	mockTokenSvc := new(MockLoginTokenService)

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("CorrectPassword123"), 12)

	mockRepo.On("GetUserByEmail", mock.Anything, "test@example.com").Return(db.User{
		ID:            userID,
		Email:         "test@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
		EmailVerified: sql.NullBool{Bool: false, Valid: true},
	}, nil)

	handler := handlers.LoginHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.LoginInput{
		Email:    "test@example.com",
		Password: "WrongPassword123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, handlers.ErrInvalidCredentials))
}
