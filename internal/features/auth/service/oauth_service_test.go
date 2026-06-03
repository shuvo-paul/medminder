package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/pkg/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOAuthAccountRepository struct {
	mock.Mock
}

func (m *MockOAuthAccountRepository) CreateOAuthAccount(ctx context.Context, id uuid.UUID, userID uuid.UUID, provider string, providerUserID string) (db.OauthAccount, error) {
	args := m.Called(ctx, id, userID, provider, providerUserID)
	if args.Get(0) == nil {
		return db.OauthAccount{}, args.Error(1)
	}
	return args.Get(0).(db.OauthAccount), args.Error(1)
}

func (m *MockOAuthAccountRepository) GetOAuthAccountByProviderAndUserID(ctx context.Context, provider string, providerUserID string) (db.OauthAccount, error) {
	args := m.Called(ctx, provider, providerUserID)
	if args.Get(0) == nil {
		return db.OauthAccount{}, args.Error(1)
	}
	return args.Get(0).(db.OauthAccount), args.Error(1)
}

func (m *MockOAuthAccountRepository) GetOAuthAccountByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) (db.OauthAccount, error) {
	args := m.Called(ctx, userID, provider)
	if args.Get(0) == nil {
		return db.OauthAccount{}, args.Error(1)
	}
	return args.Get(0).(db.OauthAccount), args.Error(1)
}

func (m *MockOAuthAccountRepository) GetOAuthAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]db.OauthAccount, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.OauthAccount), args.Error(1)
}

func (m *MockOAuthAccountRepository) DeleteOAuthAccount(ctx context.Context, id uuid.UUID) (db.OauthAccount, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return db.OauthAccount{}, args.Error(1)
	}
	return args.Get(0).(db.OauthAccount), args.Error(1)
}

func (m *MockOAuthAccountRepository) DeleteOAuthAccountByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) error {
	args := m.Called(ctx, userID, provider)
	return args.Error(0)
}

type MockOAuthUserRepository struct {
	mock.Mock
}

func (m *MockOAuthUserRepository) CreateUser(ctx context.Context, email, displayName, passwordHash string, emailVerified bool) (db.CreateUserRow, error) {
	args := m.Called(ctx, email, displayName, passwordHash, emailVerified)
	if args.Get(0) == nil {
		return db.CreateUserRow{}, args.Error(1)
	}
	return args.Get(0).(db.CreateUserRow), args.Error(1)
}

func (m *MockOAuthUserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return db.User{}, args.Error(1)
	}
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockOAuthUserRepository) GetUserByID(ctx context.Context, id string) (db.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return db.User{}, args.Error(1)
	}
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockOAuthUserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockOAuthUserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockAuditRepository is a mock for the audit log repository.
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) LogEvent(ctx context.Context, eventType string, userID uuid.NullUUID, ipAddress, userAgent string, metadata map[string]string) error {
	args := m.Called(ctx, eventType, userID, ipAddress, userAgent, metadata)
	return args.Error(0)
}

// newMockAuditRepo creates a MockAuditRepository pre-configured to accept any LogEvent call.
func newMockAuditRepo() *MockAuditRepository {
	m := new(MockAuditRepository)
	m.On("LogEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return m
}

func TestGetOrCreateUserByOAuth_NewUserRegistration(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	auditRepo := newMockAuditRepo()
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, auditRepo)

	provider := "google"
	userInfo := &oauth.UserInfo{
		ProviderUserID: "google-123",
		Email:          "newuser@example.com",
		EmailVerified:  true,
		Name:           "New User",
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	userRepo.On("GetUserByEmail", mock.Anything, userInfo.Email).
		Return(db.User{}, sql.ErrNoRows)
	userRepo.On("CreateUser", mock.Anything, userInfo.Email, userInfo.Name, "", true).
		Return(db.CreateUserRow{
			ID:            uuid.New(),
			Email:         userInfo.Email,
			DisplayName:   userInfo.Name,
			EmailVerified: sql.NullBool{Bool: true, Valid: true},
		}, nil)
	oauthRepo.On("CreateOAuthAccount", mock.Anything, mock.Anything, mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, nil)

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, userInfo)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userInfo.Email, result.Email)
	assert.Equal(t, userInfo.Name, result.DisplayName)
	assert.True(t, result.EmailVerified)
	assert.IsType(t, &service.OAuthUser{}, result)
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_connected", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetOrCreateUserByOAuth_ExistingUserLogin(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	auditRepo := newMockAuditRepo()
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, auditRepo)

	provider := "google"
	existingUserID := uuid.New()
	providerUserID := "google-456"

	oauthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         existingUserID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}
	existingUser := db.User{
		ID:            existingUserID,
		Email:         "existing@example.com",
		DisplayName:   "Existing User",
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(oauthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, existingUserID.String()).
		Return(existingUser, nil)

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, &oauth.UserInfo{
		ProviderUserID: providerUserID,
		Email:          "existing@example.com",
		EmailVerified:  true,
		Name:           "Existing User",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingUser.Email, result.Email)
	assert.Equal(t, existingUser.DisplayName, result.DisplayName)
	assert.IsType(t, &service.OAuthUser{}, result)
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_login", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetOrCreateUserByOAuth_EmailCollision(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	userInfo := &oauth.UserInfo{
		ProviderUserID: "google-789",
		Email:          "colliding@example.com",
		EmailVerified:  true,
		Name:           "New User",
	}

	existingUser := db.User{
		ID:            uuid.New(),
		Email:         "colliding@example.com",
		DisplayName:   "Existing User",
		PasswordHash:  sql.NullString{String: "hash", Valid: true},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	userRepo.On("GetUserByEmail", mock.Anything, userInfo.Email).
		Return(existingUser, nil)

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, userInfo)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, service.ErrEmailExists))
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetUserByOAuth(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	existingUserID := uuid.New()
	providerUserID := "google-123"

	oauthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         existingUserID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}
	existingUser := db.User{
		ID:            existingUserID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(oauthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, existingUserID.String()).
		Return(existingUser, nil)

	result, err := oauthSvc.GetUserByOAuth(context.Background(), provider, providerUserID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingUser.Email, result.Email)
	assert.Equal(t, existingUser.DisplayName, result.DisplayName)
	assert.IsType(t, &service.OAuthUser{}, result)
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetUserByOAuth_AccountNotFound(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	providerUserID := "unknown-user"

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)

	result, err := oauthSvc.GetUserByOAuth(context.Background(), provider, providerUserID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, repository.ErrOAuthAccountNotFound))
	oauthRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_OAuthAccountRepoError(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	userInfo := &oauth.UserInfo{
		ProviderUserID: "google-error",
		Email:          "error@example.com",
		EmailVerified:  true,
		Name:           "Error User",
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, errors.New("database connection error"))

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, userInfo)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection error")
	oauthRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_UserRepoGetByEmailError(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	userInfo := &oauth.UserInfo{
		ProviderUserID: "google-email-error",
		Email:          "email-error@example.com",
		EmailVerified:  true,
		Name:           "Email Error User",
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	userRepo.On("GetUserByEmail", mock.Anything, userInfo.Email).
		Return(db.User{}, errors.New("database error"))

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, userInfo)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_CreateUserError(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	userInfo := &oauth.UserInfo{
		ProviderUserID: "google-create-error",
		Email:          "create-error@example.com",
		EmailVerified:  true,
		Name:           "Create Error User",
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	userRepo.On("GetUserByEmail", mock.Anything, userInfo.Email).
		Return(db.User{}, sql.ErrNoRows)
	userRepo.On("CreateUser", mock.Anything, userInfo.Email, userInfo.Name, "", true).
		Return(db.CreateUserRow{}, errors.New("user creation failed"))

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, userInfo)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user creation failed")
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_CreateOAuthAccountError(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	userInfo := &oauth.UserInfo{
		ProviderUserID: "google-oauth-error",
		Email:          "oauth-error@example.com",
		EmailVerified:  true,
		Name:           "OAuth Error User",
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	userRepo.On("GetUserByEmail", mock.Anything, userInfo.Email).
		Return(db.User{}, sql.ErrNoRows)
	userRepo.On("CreateUser", mock.Anything, userInfo.Email, userInfo.Name, "", true).
		Return(db.CreateUserRow{
			ID:            uuid.New(),
			Email:         userInfo.Email,
			DisplayName:   userInfo.Name,
			EmailVerified: sql.NullBool{Bool: true, Valid: true},
		}, nil)
	oauthRepo.On("CreateOAuthAccount", mock.Anything, mock.Anything, mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, errors.New("oauth account creation failed"))
	userRepo.On("VerifyEmail", mock.Anything, mock.Anything).Return(nil)

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, userInfo)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, service.ErrOAuthProviderError))
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_CreateOAuthAccountError_VerifyEmailFails(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	userInfo := &oauth.UserInfo{
		ProviderUserID: "google-oauth-verify-fail",
		Email:          "verify-fail@example.com",
		EmailVerified:  true,
		Name:           "OAuth Verify Fail User",
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	userRepo.On("GetUserByEmail", mock.Anything, userInfo.Email).
		Return(db.User{}, sql.ErrNoRows)
	userRepo.On("CreateUser", mock.Anything, userInfo.Email, userInfo.Name, "", true).
		Return(db.CreateUserRow{
			ID:            uuid.New(),
			Email:         userInfo.Email,
			DisplayName:   userInfo.Name,
			EmailVerified: sql.NullBool{Bool: true, Valid: true},
		}, nil)
	oauthRepo.On("CreateOAuthAccount", mock.Anything, mock.Anything, mock.Anything, provider, userInfo.ProviderUserID).
		Return(db.OauthAccount{}, errors.New("oauth account creation failed"))
	userRepo.On("VerifyEmail", mock.Anything, mock.Anything).Return(errors.New("verify email failed"))

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, userInfo)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, service.ErrOAuthProviderError))
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_GetUserByIDError(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	existingUserID := uuid.New()
	providerUserID := "google-get-error"

	oauthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         existingUserID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(oauthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, existingUserID.String()).
		Return(db.User{}, errors.New("user lookup failed"))

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), provider, &oauth.UserInfo{
		ProviderUserID: providerUserID,
		Email:          "get-error@example.com",
		EmailVerified:  true,
		Name:           "Get Error User",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user lookup failed")
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetUserByOAuth_GetUserByIDError(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	provider := "google"
	existingUserID := uuid.New()
	providerUserID := "google-user-error"

	oauthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         existingUserID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(oauthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, existingUserID.String()).
		Return(db.User{}, errors.New("user not found"))

	result, err := oauthSvc.GetUserByOAuth(context.Background(), provider, providerUserID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_NilUserInfo(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	result, err := oauthSvc.GetOrCreateUserByOAuth(context.Background(), "google", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, service.ErrOAuthProviderError))
}

func TestLinkOAuthAccount_Success(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	auditRepo := newMockAuditRepo()
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, auditRepo)

	userID := uuid.New()
	provider := "google"
	providerUserID := "google-123"

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "somehash", Valid: true},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{}, nil)
	oauthRepo.On("CreateOAuthAccount", mock.Anything, mock.Anything, userID, provider, providerUserID).
		Return(db.OauthAccount{}, nil)

	err := oauthSvc.LinkOAuthAccount(context.Background(), userID, provider, providerUserID)

	assert.NoError(t, err)
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_linked", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestLinkOAuthAccount_UpsertReplacesOldLink(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	auditRepo := newMockAuditRepo()
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, auditRepo)

	userID := uuid.New()
	provider := "google"
	oldProviderUserID := "google-old"
	newProviderUserID := "google-new"

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "somehash", Valid: true},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	existingOAuthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: oldProviderUserID,
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, newProviderUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(existingOAuthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{existingOAuthAccount}, nil)
	oauthRepo.On("DeleteOAuthAccount", mock.Anything, existingOAuthAccount.ID).
		Return(existingOAuthAccount, nil)
	oauthRepo.On("CreateOAuthAccount", mock.Anything, mock.Anything, userID, provider, newProviderUserID).
		Return(db.OauthAccount{}, nil)

	err := oauthSvc.LinkOAuthAccount(context.Background(), userID, provider, newProviderUserID)

	assert.NoError(t, err)
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_linked", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestLinkOAuthAccount_DeadEndPrevention_NoPasswordSingleProvider(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()
	provider := "google"
	providerUserID := "google-123"

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "", Valid: false},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	existingOAuthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(existingOAuthAccount, nil)
	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(existingOAuthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{existingOAuthAccount}, nil)

	err := oauthSvc.LinkOAuthAccount(context.Background(), userID, provider, providerUserID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrAccountWillBeLocked))
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestLinkOAuthAccount_DeadEndPrevention_NewLinkWouldLeaveNoPassword(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()
	provider := "google"
	providerUserID := "google-new"

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "", Valid: false},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{}, nil)

	err := oauthSvc.LinkOAuthAccount(context.Background(), userID, provider, providerUserID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrAccountWillBeLocked))
}

func TestLinkOAuthAccount_ProviderAlreadyLinkedToOtherUser(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()
	differentUserID := uuid.New()
	provider := "google"
	providerUserID := "google-123"

	ownedAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         differentUserID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}

	oauthRepo.On("GetOAuthAccountByProviderAndUserID", mock.Anything, provider, providerUserID).
		Return(ownedAccount, nil)

	err := oauthSvc.LinkOAuthAccount(context.Background(), userID, provider, providerUserID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrProviderAlreadyLinked))
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestUnlinkOAuthAccount_Success(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	auditRepo := newMockAuditRepo()
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, auditRepo)

	userID := uuid.New()
	provider := "google"
	providerUserID := "google-123"

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "somehash", Valid: true},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}

	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(oauthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{oauthAccount}, nil)
	oauthRepo.On("DeleteOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(nil)

	err := oauthSvc.UnlinkOAuthAccount(context.Background(), userID, provider)

	assert.NoError(t, err)
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_unlinked", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestUnlinkOAuthAccount_AccountNotFound(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()
	provider := "google"

	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(db.OauthAccount{}, repository.ErrOAuthAccountNotFound)

	err := oauthSvc.UnlinkOAuthAccount(context.Background(), userID, provider)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrOAuthAccountNotFound))
	oauthRepo.AssertExpectations(t)
}

func TestUnlinkOAuthAccount_DeadEndPrevention_NoPassword(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()
	provider := "google"
	providerUserID := "google-123"

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "", Valid: false},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}

	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(oauthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{oauthAccount}, nil)

	err := oauthSvc.UnlinkOAuthAccount(context.Background(), userID, provider)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrAccountWillBeLocked))
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestUnlinkOAuthAccount_DeadEndPrevention_OnlyProviderNoPassword(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()
	provider := "google"
	providerUserID := "google-123"

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "", Valid: false},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}

	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(oauthAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{oauthAccount}, nil)

	err := oauthSvc.UnlinkOAuthAccount(context.Background(), userID, provider)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrAccountWillBeLocked))
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestUnlinkOAuthAccount_MultiProviderNoPassword(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	auditRepo := newMockAuditRepo()
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, auditRepo)

	userID := uuid.New()
	provider := "google"

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "", Valid: false},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	googleAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       "google",
		ProviderUserID: "google-123",
	}
	githubAccount := db.OauthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       "github",
		ProviderUserID: "github-456",
	}

	oauthRepo.On("GetOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(googleAccount, nil)
	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{googleAccount, githubAccount}, nil)
	oauthRepo.On("DeleteOAuthAccountByUserIDAndProvider", mock.Anything, userID, provider).
		Return(nil)

	err := oauthSvc.UnlinkOAuthAccount(context.Background(), userID, provider)

	assert.NoError(t, err)
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	auditRepo.AssertCalled(t, "LogEvent", mock.Anything, "oauth_unlinked", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGetUserOAuthProviders_Success(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()

	oauthAccounts := []db.OauthAccount{
		{ID: uuid.New(), UserID: userID, Provider: "google", ProviderUserID: "google-1"},
		{ID: uuid.New(), UserID: userID, Provider: "github", ProviderUserID: "github-1"},
	}

	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return(oauthAccounts, nil)

	providers, err := oauthSvc.GetUserOAuthProviders(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, []string{"google", "github"}, providers)
	oauthRepo.AssertExpectations(t)
}

func TestGetUserOAuthProviders_Empty(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()

	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return([]db.OauthAccount{}, nil)

	providers, err := oauthSvc.GetUserOAuthProviders(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, []string{}, providers)
	oauthRepo.AssertExpectations(t)
}

func TestCanUnlinkProvider_HasPassword(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "somehash", Valid: true},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)

	canUnlink, err := oauthSvc.CanUnlinkProvider(context.Background(), userID, "google")

	assert.NoError(t, err)
	assert.True(t, canUnlink)
	userRepo.AssertExpectations(t)
}

func TestCanUnlinkProvider_NoPasswordMultipleProviders(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "", Valid: false},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthAccounts := []db.OauthAccount{
		{ID: uuid.New(), UserID: userID, Provider: "google", ProviderUserID: "google-1"},
		{ID: uuid.New(), UserID: userID, Provider: "github", ProviderUserID: "github-1"},
	}

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return(oauthAccounts, nil)

	canUnlink, err := oauthSvc.CanUnlinkProvider(context.Background(), userID, "google")

	assert.NoError(t, err)
	assert.True(t, canUnlink)
	userRepo.AssertExpectations(t)
	oauthRepo.AssertExpectations(t)
}

func TestCanUnlinkProvider_NoPasswordSingleProvider(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo, newMockAuditRepo())

	userID := uuid.New()

	existingUser := db.User{
		ID:            userID,
		Email:         "user@example.com",
		DisplayName:   "Test User",
		PasswordHash:  sql.NullString{String: "", Valid: false},
		EmailVerified: sql.NullBool{Bool: true, Valid: true},
	}

	oauthAccounts := []db.OauthAccount{
		{ID: uuid.New(), UserID: userID, Provider: "google", ProviderUserID: "google-1"},
	}

	userRepo.On("GetUserByID", mock.Anything, userID.String()).
		Return(existingUser, nil)
	oauthRepo.On("GetOAuthAccountsByUserID", mock.Anything, userID).
		Return(oauthAccounts, nil)

	canUnlink, err := oauthSvc.CanUnlinkProvider(context.Background(), userID, "google")

	assert.NoError(t, err)
	assert.False(t, canUnlink)
	userRepo.AssertExpectations(t)
	oauthRepo.AssertExpectations(t)
}
