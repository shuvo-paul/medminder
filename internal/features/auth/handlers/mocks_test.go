package handlers_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
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

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error) {
	args := m.Called(ctx, userID, tokenHash, expiresAt)
	return args.Get(0).(db.CreateRefreshTokenRow), args.Error(1)
}

func (m *MockRefreshTokenRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(db.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteRefreshToken(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
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

type MockRegisterUserRepository struct {
	mock.Mock
}

func (m *MockRegisterUserRepository) CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error) {
	args := m.Called(ctx, email, displayName, passwordHash)
	if args.Get(0) == nil {
		return db.CreateUserRow{}, args.Error(1)
	}
	return args.Get(0).(db.CreateUserRow), args.Error(1)
}

func (m *MockRegisterUserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockRegisterUserRepository) GetUserByID(ctx context.Context, id string) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockRegisterUserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}
