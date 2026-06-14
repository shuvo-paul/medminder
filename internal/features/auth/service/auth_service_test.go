package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository mocks repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, email, displayName, passwordHash string, emailVerified bool) (db.CreateUserRow, error) {
	args := m.Called(ctx, email, displayName, passwordHash, emailVerified)
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

func (m *MockUserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserEmail(ctx context.Context, id uuid.UUID, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockRefreshTokenRepository mocks repository.RefreshTokenRepository
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

// MockTokenService mocks service.TokenServiceInterface
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

// Test cases for Register
func TestRegister(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*MockUserRepository, *MockTokenService)
		input  struct{ email, displayName, password string }
		expect func(*testing.T, *service.RegisterResult, error)
	}{
		{
			name: "Register_Success",
			setup: func(userRepo *MockUserRepository, tokenSvc *MockTokenService) {
				userRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{}, sql.ErrNoRows)
				userRepo.On("CreateUser", mock.Anything, "test@example.com", "Test User", mock.AnythingOfType("string"), false).
					Return(db.CreateUserRow{
						ID:            uuid.New(),
						Email:         "test@example.com",
						DisplayName:   "Test User",
						EmailVerified: sql.NullBool{Bool: false, Valid: true},
					}, nil)
			},
			input: struct{ email, displayName, password string }{
				email:       "test@example.com",
				displayName: "Test User",
				password:    "password123",
			},
			expect: func(t *testing.T, result *service.RegisterResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test@example.com", result.User.Email)
				assert.Equal(t, "Test User", result.User.DisplayName)
				assert.False(t, result.User.EmailVerified)
			},
		},
		{
			name: "Register_EmailExists",
			setup: func(userRepo *MockUserRepository, tokenSvc *MockTokenService) {
				userRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{
						ID:            uuid.New(),
						Email:         "test@example.com",
						DisplayName:   "Existing User",
						PasswordHash:  sql.NullString{String: "hash", Valid: true},
						EmailVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil)
			},
			input: struct{ email, displayName, password string }{
				email:       "test@example.com",
				displayName: "Test User",
				password:    "password123",
			},
			expect: func(t *testing.T, result *service.RegisterResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.True(t, errors.Is(err, service.ErrEmailExists))
			},
		},
		{
			name: "Register_CreateFails",
			setup: func(userRepo *MockUserRepository, tokenSvc *MockTokenService) {
				userRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{}, sql.ErrNoRows)
				userRepo.On("CreateUser", mock.Anything, "test@example.com", "Test User", mock.AnythingOfType("string"), false).
					Return(db.CreateUserRow{}, errors.New("database error"))
			},
			input: struct{ email, displayName, password string }{
				email:       "test@example.com",
				displayName: "Test User",
				password:    "password123",
			},
			expect: func(t *testing.T, result *service.RegisterResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockRefreshTokenRepository)
			tokenSvc := new(MockTokenService)
			authSvc := service.NewAuthService(userRepo, tokenRepo, tokenSvc)

			tt.setup(userRepo, tokenSvc)

			// Act
			result, err := authSvc.Register(context.Background(), tt.input.email, tt.input.displayName, tt.input.password)

			// Assert
			tt.expect(t, result, err)
			userRepo.AssertExpectations(t)
		})
	}
}

// Test cases for Login
func TestLogin(t *testing.T) {
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), service.BcryptCost)

	tests := []struct {
		name   string
		setup  func(*MockUserRepository, *MockRefreshTokenRepository, *MockTokenService)
		input  struct{ email, password string }
		expect func(*testing.T, *service.LoginResult, error)
	}{
		{
			name: "Login_Success",
			setup: func(userRepo *MockUserRepository, tokenRepo *MockRefreshTokenRepository, tokenSvc *MockTokenService) {
				userID := uuid.New()
				userRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{
						ID:            userID,
						Email:         "test@example.com",
						DisplayName:   "Test User",
						PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
						EmailVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil)
				tokenSvc.On("GenerateAccessToken", userID, "test@example.com").
					Return("access_token", nil)
				tokenSvc.On("GenerateRefreshToken").
					Return("refresh_token", nil)
				tokenSvc.On("HashRefreshToken", "refresh_token").
					Return("token_hash")
				tokenRepo.On("CreateRefreshToken", mock.Anything, userID, "token_hash", mock.AnythingOfType("time.Time")).
					Return(db.CreateRefreshTokenRow{}, nil)
			},
			input: struct{ email, password string }{
				email:    "test@example.com",
				password: password,
			},
			expect: func(t *testing.T, result *service.LoginResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "access_token", result.AccessToken)
				assert.Equal(t, "refresh_token", result.RefreshToken)
				assert.Equal(t, "test@example.com", result.User.Email)
			},
		},
		{
			name: "Login_UserNotFound",
			setup: func(userRepo *MockUserRepository, tokenRepo *MockRefreshTokenRepository, tokenSvc *MockTokenService) {
				userRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{}, sql.ErrNoRows)
			},
			input: struct{ email, password string }{
				email:    "test@example.com",
				password: password,
			},
			expect: func(t *testing.T, result *service.LoginResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.True(t, errors.Is(err, service.ErrInvalidCredentials))
			},
		},
		{
			name: "Login_EmptyPasswordHash",
			setup: func(userRepo *MockUserRepository, tokenRepo *MockRefreshTokenRepository, tokenSvc *MockTokenService) {
				userRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{
						ID:            uuid.New(),
						Email:         "test@example.com",
						DisplayName:   "Test User",
						PasswordHash:  sql.NullString{Valid: false},
						EmailVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil)
			},
			input: struct{ email, password string }{
				email:    "test@example.com",
				password: password,
			},
			expect: func(t *testing.T, result *service.LoginResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.True(t, errors.Is(err, service.ErrInvalidCredentials))
			},
		},
		{
			name: "Login_WrongPassword",
			setup: func(userRepo *MockUserRepository, tokenRepo *MockRefreshTokenRepository, tokenSvc *MockTokenService) {
				userRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{
						ID:            uuid.New(),
						Email:         "test@example.com",
						DisplayName:   "Test User",
						PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
						EmailVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil)
			},
			input: struct{ email, password string }{
				email:    "test@example.com",
				password: "wrongpassword",
			},
			expect: func(t *testing.T, result *service.LoginResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.True(t, errors.Is(err, service.ErrInvalidCredentials))
			},
		},
		{
			name: "Login_TokenGenerationFails",
			setup: func(userRepo *MockUserRepository, tokenRepo *MockRefreshTokenRepository, tokenSvc *MockTokenService) {
				userID := uuid.New()
				userRepo.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(db.User{
						ID:            userID,
						Email:         "test@example.com",
						DisplayName:   "Test User",
						PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
						EmailVerified: sql.NullBool{Bool: true, Valid: true},
					}, nil)
				tokenSvc.On("GenerateAccessToken", userID, "test@example.com").
					Return("", errors.New("token generation failed"))
			},
			input: struct{ email, password string }{
				email:    "test@example.com",
				password: password,
			},
			expect: func(t *testing.T, result *service.LoginResult, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "token generation failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockRefreshTokenRepository)
			tokenSvc := new(MockTokenService)
			authSvc := service.NewAuthService(userRepo, tokenRepo, tokenSvc)

			tt.setup(userRepo, tokenRepo, tokenSvc)

			// Act
			result, err := authSvc.Login(context.Background(), tt.input.email, tt.input.password)

			// Assert
			tt.expect(t, result, err)
			userRepo.AssertExpectations(t)
			tokenRepo.AssertExpectations(t)
			tokenSvc.AssertExpectations(t)
		})
	}
}

// Test cases for Logout
func TestLogout(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name   string
		setup  func(*MockRefreshTokenRepository, uuid.UUID)
		input  struct{ userID uuid.UUID }
		expect func(*testing.T, error)
	}{
		{
			name: "Logout_Success",
			setup: func(tokenRepo *MockRefreshTokenRepository, userID uuid.UUID) {
				tokenRepo.On("DeleteUserRefreshTokens", mock.Anything, userID).
					Return(nil)
			},
			input: struct{ userID uuid.UUID }{userID: userID},
			expect: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "Logout_NilUserID",
			setup: func(tokenRepo *MockRefreshTokenRepository, userID uuid.UUID) {
				// No mock setup needed - should fail before calling repo
			},
			input: struct{ userID uuid.UUID }{userID: uuid.Nil},
			expect: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrLogoutFailed))
			},
		},
		{
			name: "Logout_DBError",
			setup: func(tokenRepo *MockRefreshTokenRepository, userID uuid.UUID) {
				tokenRepo.On("DeleteUserRefreshTokens", mock.Anything, userID).
					Return(errors.New("database error"))
			},
			input: struct{ userID uuid.UUID }{userID: userID},
			expect: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrLogoutFailed))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockRefreshTokenRepository)
			tokenSvc := new(MockTokenService)
			authSvc := service.NewAuthService(userRepo, tokenRepo, tokenSvc)

			tt.setup(tokenRepo, tt.input.userID)

			// Act
			err := authSvc.Logout(context.Background(), tt.input.userID)

			// Assert
			tt.expect(t, err)
			tokenRepo.AssertExpectations(t)
		})
	}
}

func TestChangePassword(t *testing.T) {
	currentPwd := "CurrentPwd1"
	hashedCurrent, _ := bcrypt.GenerateFromPassword([]byte(currentPwd), service.BcryptCost)
	newPwd := "NewPwd123"

	type changePwdInput struct {
		userID          uuid.UUID
		currentPassword string
		newPassword     string
	}

	tests := []struct {
		name   string
		input  changePwdInput
		setup  func(*MockUserRepository, changePwdInput)
		expect func(*testing.T, error)
	}{
		{
			name: "ChangePassword_Success",
			input: changePwdInput{
				userID:          uuid.New(),
				currentPassword: currentPwd,
				newPassword:     newPwd,
			},
			setup: func(userRepo *MockUserRepository, input changePwdInput) {
				userRepo.On("GetUserByID", mock.Anything, input.userID.String()).
					Return(db.User{
						ID:           input.userID,
						PasswordHash: sql.NullString{String: string(hashedCurrent), Valid: true},
					}, nil)
				userRepo.On("UpdatePassword", mock.Anything, input.userID.String(), mock.AnythingOfType("string")).
					Return(nil)
			},
			expect: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "ChangePassword_UserNotFound",
			input: changePwdInput{
				userID:          uuid.New(),
				currentPassword: currentPwd,
				newPassword:     newPwd,
			},
			setup: func(userRepo *MockUserRepository, input changePwdInput) {
				userRepo.On("GetUserByID", mock.Anything, input.userID.String()).
					Return(db.User{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrUserNotFound))
			},
		},
		{
			name: "ChangePassword_NoPasswordSet",
			input: changePwdInput{
				userID:          uuid.New(),
				currentPassword: currentPwd,
				newPassword:     newPwd,
			},
			setup: func(userRepo *MockUserRepository, input changePwdInput) {
				userRepo.On("GetUserByID", mock.Anything, input.userID.String()).
					Return(db.User{
						ID:           input.userID,
						PasswordHash: sql.NullString{Valid: false},
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrNoPasswordSet))
			},
		},
		{
			name: "ChangePassword_WrongCurrentPassword",
			input: changePwdInput{
				userID:          uuid.New(),
				currentPassword: "WrongPwd1",
				newPassword:     newPwd,
			},
			setup: func(userRepo *MockUserRepository, input changePwdInput) {
				userRepo.On("GetUserByID", mock.Anything, input.userID.String()).
					Return(db.User{
						ID:           input.userID,
						PasswordHash: sql.NullString{String: string(hashedCurrent), Valid: true},
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrWrongPassword))
			},
		},
		{
			name: "ChangePassword_SamePassword",
			input: changePwdInput{
				userID:          uuid.New(),
				currentPassword: currentPwd,
				newPassword:     currentPwd,
			},
			setup: func(userRepo *MockUserRepository, input changePwdInput) {
				userRepo.On("GetUserByID", mock.Anything, input.userID.String()).
					Return(db.User{
						ID:           input.userID,
						PasswordHash: sql.NullString{String: string(hashedCurrent), Valid: true},
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, service.ErrSamePassword))
			},
		},
		{
			name: "ChangePassword_UpdateFails",
			input: changePwdInput{
				userID:          uuid.New(),
				currentPassword: currentPwd,
				newPassword:     newPwd,
			},
			setup: func(userRepo *MockUserRepository, input changePwdInput) {
				userRepo.On("GetUserByID", mock.Anything, input.userID.String()).
					Return(db.User{
						ID:           input.userID,
						PasswordHash: sql.NullString{String: string(hashedCurrent), Valid: true},
					}, nil)
				userRepo.On("UpdatePassword", mock.Anything, input.userID.String(), mock.AnythingOfType("string")).
					Return(errors.New("database error"))
			},
			expect: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "updating password")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			tokenRepo := new(MockRefreshTokenRepository)
			tokenSvc := new(MockTokenService)
			authSvc := service.NewAuthService(userRepo, tokenRepo, tokenSvc)

			tt.setup(userRepo, tt.input)

			err := authSvc.ChangePassword(context.Background(), tt.input.userID, tt.input.currentPassword, tt.input.newPassword)

			tt.expect(t, err)
			userRepo.AssertExpectations(t)
		})
	}
}
