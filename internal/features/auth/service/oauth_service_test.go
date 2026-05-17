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

func TestGetOrCreateUserByOAuth_NewUserRegistration(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_ExistingUserLogin(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetOrCreateUserByOAuth_EmailCollision(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetUserByOAuth_AccountNotFound(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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

func TestGetOrCreateUserByOAuth_GetUserByIDError(t *testing.T) {
	userRepo := new(MockOAuthUserRepository)
	oauthRepo := new(MockOAuthAccountRepository)
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
	oauthSvc := service.NewOAuthService(userRepo, oauthRepo)

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
