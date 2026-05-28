package handlers_test

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	db "github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/pkg/oauth"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, displayName, password string) (*service.RegisterResult, error) {
	args := m.Called(ctx, email, displayName, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RegisterResult), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*service.LoginResult, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResult), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
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

func (m *MockTokenService) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

type MockEmailVerificationService struct {
	mock.Mock
}

func (m *MockEmailVerificationService) VerifyEmail(ctx context.Context, token string) (*service.VerifyResult, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.VerifyResult), args.Error(1)
}

func (m *MockEmailVerificationService) ResendVerification(ctx context.Context, userID uuid.UUID, email string) error {
	args := m.Called(ctx, userID, email)
	return args.Error(0)
}

// -- OAuth mocks --

type MockOAuthService struct {
	mock.Mock
}

func (m *MockOAuthService) GetOrCreateUserByOAuth(ctx context.Context, provider string, userInfo *oauth.UserInfo) (*service.OAuthUser, error) {
	args := m.Called(ctx, provider, userInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.OAuthUser), args.Error(1)
}

func (m *MockOAuthService) GetUserByOAuth(ctx context.Context, provider, providerUserID string) (*service.OAuthUser, error) {
	args := m.Called(ctx, provider, providerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.OAuthUser), args.Error(1)
}

func (m *MockOAuthService) LinkOAuthAccount(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error {
	args := m.Called(ctx, userID, provider, providerUserID)
	return args.Error(0)
}

func (m *MockOAuthService) UnlinkOAuthAccount(ctx context.Context, userID uuid.UUID, provider string) error {
	args := m.Called(ctx, userID, provider)
	return args.Error(0)
}

func (m *MockOAuthService) GetUserOAuthProviders(ctx context.Context, userID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockOAuthService) CanUnlinkProvider(ctx context.Context, userID uuid.UUID, provider string) (bool, error) {
	args := m.Called(ctx, userID, provider)
	return args.Bool(0), args.Error(1)
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

type MockOAuthAuthorizationCodeRepository struct {
	mock.Mock
}

func (m *MockOAuthAuthorizationCodeRepository) CreateAuthorizationCode(ctx context.Context, id uuid.UUID, codeHash string, userID uuid.NullUUID, nonce string, purpose string, expiresAt time.Time) (db.OauthAuthorizationCode, error) {
	args := m.Called(ctx, id, codeHash, userID, nonce, purpose, expiresAt)
	return args.Get(0).(db.OauthAuthorizationCode), args.Error(1)
}

func (m *MockOAuthAuthorizationCodeRepository) CreateAuthorizationCodeWithUserInfo(ctx context.Context, id uuid.UUID, codeHash string, nonce string, purpose string, expiresAt time.Time, provider string, providerUserID string, providerEmail string, providerName string, providerEmailVerified bool) (db.OauthAuthorizationCode, error) {
	args := m.Called(ctx, id, codeHash, nonce, purpose, expiresAt, provider, providerUserID, providerEmail, providerName, providerEmailVerified)
	return args.Get(0).(db.OauthAuthorizationCode), args.Error(1)
}

func (m *MockOAuthAuthorizationCodeRepository) GetAuthorizationCodeByHash(ctx context.Context, codeHash string) (db.OauthAuthorizationCode, error) {
	args := m.Called(ctx, codeHash)
	return args.Get(0).(db.OauthAuthorizationCode), args.Error(1)
}

func (m *MockOAuthAuthorizationCodeRepository) GetAndLockAuthorizationCode(ctx context.Context, codeHash string) (*repository.AuthorizationCodeInfo, error) {
	args := m.Called(ctx, codeHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AuthorizationCodeInfo), args.Error(1)
}

func (m *MockOAuthAuthorizationCodeRepository) MarkAuthorizationCodeAsUsed(ctx context.Context, codeHash string) (db.OauthAuthorizationCode, error) {
	args := m.Called(ctx, codeHash)
	return args.Get(0).(db.OauthAuthorizationCode), args.Error(1)
}

func (m *MockOAuthAuthorizationCodeRepository) CleanupExpiredAuthorizationCodes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
