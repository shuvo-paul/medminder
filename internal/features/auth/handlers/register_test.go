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
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error) {
	args := m.Called(ctx, email, displayName, passwordHash)
	if args.Get(0) == nil {
		return db.CreateUserRow{}, args.Error(1)
	}
	return args.Get(0).(db.CreateUserRow), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error) {
	args := m.Called(ctx, userID, tokenHash, expiresAt)
	return args.Get(0).(db.CreateRefreshTokenRow), args.Error(1)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GenerateRefreshToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) HashRefreshToken(token string) string {
	args := m.Called(token)
	return args.String(0)
}

func TestRegister_Successful(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenSvc := new(MockTokenService)

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
	mockTokenSvc.On("GenerateAccessToken", userID, email).Return("access-token", nil)
	mockTokenSvc.On("GenerateRefreshToken").Return("refresh-token", nil)
	mockTokenSvc.On("HashRefreshToken", "refresh-token").Return("hashed-token")
	mockRepo.On("CreateRefreshToken", mock.Anything, userID, "hashed-token", mock.Anything).Return(db.CreateRefreshTokenRow{}, nil)

	handler := handlers.RegisterHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       email,
		DisplayName: displayName,
		Password:    password,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, userID, resp.User.ID)
	assert.Equal(t, email, resp.User.Email)
	assert.Equal(t, displayName, resp.User.DisplayName)
	assert.False(t, resp.User.EmailVerified)
}

func TestRegister_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenSvc := new(MockTokenService)
	handler := handlers.RegisterHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       "invalid-email",
		DisplayName: "Test User",
		Password:    "Password123",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRegister_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenSvc := new(MockTokenService)
	handler := handlers.RegisterHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       "test@example.com",
		DisplayName: "Test User",
		Password:    "weak",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockTokenSvc := new(MockTokenService)

	email := "test@example.com"
	userID := uuid.New()

	mockRepo.On("GetUserByEmail", mock.Anything, email).Return(db.User{
		ID:    userID,
		Email: email,
	}, nil)

	handler := handlers.RegisterHandler(mockRepo, mockTokenSvc)

	resp, err := handler(context.Background(), &handlers.RegisterInput{
		Email:       email,
		DisplayName: "Test User",
		Password:    "Password123",
	})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, handlers.ErrEmailExists))
	assert.Nil(t, resp)
}
