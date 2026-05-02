package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/stretchr/testify/assert"
)

// Mock implementations

type MockPRUserRepository struct {
	GetUserByEmailFn func(ctx context.Context, email string) (db.User, error)
	GetUserByIDFn    func(ctx context.Context, id string) (db.User, error)
	UpdatePasswordFn func(ctx context.Context, id, passwordHash string) error
	CreateUserFn     func(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error)
	VerifyEmailFn     func(ctx context.Context, id uuid.UUID) error
}

func (m *MockPRUserRepository) CreateUser(ctx context.Context, email, displayName, passwordHash string) (db.CreateUserRow, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, email, displayName, passwordHash)
	}
	return db.CreateUserRow{}, nil
}

func (m *MockPRUserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(ctx, email)
	}
	return db.User{}, nil
}

func (m *MockPRUserRepository) GetUserByID(ctx context.Context, id string) (db.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return db.User{}, nil
}

func (m *MockPRUserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(ctx, id, passwordHash)
	}
	return nil
}

func (m *MockPRUserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	if m.VerifyEmailFn != nil {
		return m.VerifyEmailFn(ctx, id)
	}
	return nil
}

// MockPRPasswordResetTokenRepository

type MockPRPasswordResetTokenRepository struct {
	CreateTokenFn      func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.PasswordResetToken, error)
	FindValidTokenFn   func(ctx context.Context, tokenHash string) (db.PasswordResetToken, error)
	MarkAsUsedFn       func(ctx context.Context, tokenID uuid.UUID) error
	DeleteAllForUserFn func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockPRPasswordResetTokenRepository) CreateToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.PasswordResetToken, error) {
	if m.CreateTokenFn != nil {
		return m.CreateTokenFn(ctx, userID, tokenHash, expiresAt)
	}
	return db.PasswordResetToken{}, nil
}

func (m *MockPRPasswordResetTokenRepository) FindValidToken(ctx context.Context, tokenHash string) (db.PasswordResetToken, error) {
	if m.FindValidTokenFn != nil {
		return m.FindValidTokenFn(ctx, tokenHash)
	}
	return db.PasswordResetToken{}, nil
}

func (m *MockPRPasswordResetTokenRepository) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
	if m.MarkAsUsedFn != nil {
		return m.MarkAsUsedFn(ctx, tokenID)
	}
	return nil
}

func (m *MockPRPasswordResetTokenRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteAllForUserFn != nil {
		return m.DeleteAllForUserFn(ctx, userID)
	}
	return nil
}

// MockPRRefreshTokenRepository

type MockPRRefreshTokenRepository struct {
	DeleteAllForUserFn func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockPRRefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteAllForUserFn != nil {
		return m.DeleteAllForUserFn(ctx, userID)
	}
	return nil
}

// Unused methods to satisfy interface - not called in password reset tests
func (m *MockPRRefreshTokenRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error) {
	return db.CreateRefreshTokenRow{}, nil
}
func (m *MockPRRefreshTokenRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (m *MockPRRefreshTokenRepository) DeleteRefreshToken(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockPRRefreshTokenRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// MockPREmailClient

type MockPREmailClient struct {
	SendEmailFn func(ctx context.Context, to, subject, body string) error
}

func (m *MockPREmailClient) SendEmail(ctx context.Context, to, subject, body string) error {
	if m.SendEmailFn != nil {
		return m.SendEmailFn(ctx, to, subject, body)
	}
	return nil
}

// Tests

func TestPasswordResetService_RequestReset(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockPRUserRepository, *MockPRPasswordResetTokenRepository, *MockPREmailClient)
		expectedErr error
	}{
		{
			name: "RequestReset_UserNotFound",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, emailClient *MockPREmailClient) {
				userRepo.GetUserByEmailFn = func(ctx context.Context, email string) (db.User, error) {
					return db.User{}, sql.ErrNoRows
				}
			},
			expectedErr: nil,
		},
		{
			name: "RequestReset_Success",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, emailClient *MockPREmailClient) {
				userID := uuid.New()
				userRepo.GetUserByEmailFn = func(ctx context.Context, email string) (db.User, error) {
					return db.User{
						ID:    userID,
						Email: "test@example.com",
					}, nil
				}
				tokenRepo.DeleteAllForUserFn = func(ctx context.Context, userID uuid.UUID) error {
					return nil
				}
				tokenRepo.CreateTokenFn = func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{
						ID:        uuid.New(),
						UserID:    userID,
						TokenHash: tokenHash,
						ExpiresAt: expiresAt,
					}, nil
				}
				emailClient.SendEmailFn = func(ctx context.Context, to, subject, body string) error {
					return nil
				}
			},
			expectedErr: nil,
		},
		{
			name: "RequestReset_EmailFails",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, emailClient *MockPREmailClient) {
				userID := uuid.New()
				userRepo.GetUserByEmailFn = func(ctx context.Context, email string) (db.User, error) {
					return db.User{
						ID:    userID,
						Email: "test@example.com",
					}, nil
				}
				// First DeleteAllForUser call (cleanup before create)
				deleteCallCount := 0
				tokenRepo.DeleteAllForUserFn = func(ctx context.Context, userID uuid.UUID) error {
					deleteCallCount++
					return nil
				}
				tokenRepo.CreateTokenFn = func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{
						ID:        uuid.New(),
						UserID:    userID,
						TokenHash: tokenHash,
						ExpiresAt: expiresAt,
					}, nil
				}
				emailClient.SendEmailFn = func(ctx context.Context, to, subject, body string) error {
					return errors.New("email send failed")
				}
				// Override DeleteAllForUser to handle both pre-create and post-email-failure cleanup
				tokenRepo.DeleteAllForUserFn = func(ctx context.Context, userID uuid.UUID) error {
					return nil
				}
			},
			expectedErr: nil,
		},
		{
			name: "RequestReset_TokenCreateFails",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, emailClient *MockPREmailClient) {
				userID := uuid.New()
				userRepo.GetUserByEmailFn = func(ctx context.Context, email string) (db.User, error) {
					return db.User{
						ID:    userID,
						Email: "test@example.com",
					}, nil
				}
				tokenRepo.DeleteAllForUserFn = func(ctx context.Context, userID uuid.UUID) error {
					return nil
				}
				tokenRepo.CreateTokenFn = func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{}, errors.New("database error")
				}
			},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userRepo := &MockPRUserRepository{}
			tokenRepo := &MockPRPasswordResetTokenRepository{}
			emailClient := &MockPREmailClient{}

			if tt.setup != nil {
				tt.setup(userRepo, tokenRepo, emailClient)
			}

			svc := service.NewPasswordResetService(
				userRepo,
				tokenRepo,
				&MockPRRefreshTokenRepository{},
				emailClient,
				"https://example.com",
			)

			// Act
			err := svc.RequestReset(context.Background(), "test@example.com")

			// Assert
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordResetService_ConfirmReset(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*MockPRUserRepository, *MockPRPasswordResetTokenRepository, *MockPRRefreshTokenRepository)
		expectedErr error
	}{
		{
			name: "ConfirmReset_Success",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, refreshTokenRepo *MockPRRefreshTokenRepository) {
				tokenID := uuid.New()
				userID := uuid.New()
				tokenRepo.FindValidTokenFn = func(ctx context.Context, tokenHash string) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{
						ID:        tokenID,
						UserID:    userID,
						TokenHash: tokenHash,
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil
				}
				userRepo.GetUserByIDFn = func(ctx context.Context, id string) (db.User, error) {
					return db.User{
						ID:           userID,
						Email:        "test@example.com",
						PasswordHash: sql.NullString{String: "", Valid: false},
					}, nil
				}
				tokenRepo.MarkAsUsedFn = func(ctx context.Context, tokenID uuid.UUID) error {
					return nil
				}
				userRepo.UpdatePasswordFn = func(ctx context.Context, id, passwordHash string) error {
					return nil
				}
				refreshTokenRepo.DeleteAllForUserFn = func(ctx context.Context, userID uuid.UUID) error {
					return nil
				}
			},
			expectedErr: nil,
		},
		{
			name: "ConfirmReset_InvalidToken",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, refreshTokenRepo *MockPRRefreshTokenRepository) {
				tokenRepo.FindValidTokenFn = func(ctx context.Context, tokenHash string) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{}, repository.ErrTokenNotFound
				}
			},
			expectedErr: repository.ErrTokenNotFound,
		},
		{
			name: "ConfirmReset_ExpiredToken",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, refreshTokenRepo *MockPRRefreshTokenRepository) {
				tokenRepo.FindValidTokenFn = func(ctx context.Context, tokenHash string) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{}, repository.ErrTokenExpired
				}
			},
			expectedErr: repository.ErrTokenExpired,
		},
		{
			name: "ConfirmReset_UsedToken",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, refreshTokenRepo *MockPRRefreshTokenRepository) {
				tokenRepo.FindValidTokenFn = func(ctx context.Context, tokenHash string) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{}, repository.ErrTokenUsed
				}
			},
			expectedErr: repository.ErrTokenUsed,
		},
		{
			name: "ConfirmReset_UserNotFound",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, refreshTokenRepo *MockPRRefreshTokenRepository) {
				tokenRepo.FindValidTokenFn = func(ctx context.Context, tokenHash string) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{
						ID:        uuid.New(),
						UserID:    uuid.New(),
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil
				}
				userRepo.GetUserByIDFn = func(ctx context.Context, id string) (db.User, error) {
					return db.User{}, sql.ErrNoRows
				}
			},
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "ConfirmReset_PasswordUpdateFails",
			setup: func(userRepo *MockPRUserRepository, tokenRepo *MockPRPasswordResetTokenRepository, refreshTokenRepo *MockPRRefreshTokenRepository) {
				tokenRepo.FindValidTokenFn = func(ctx context.Context, tokenHash string) (db.PasswordResetToken, error) {
					return db.PasswordResetToken{
						ID:        uuid.New(),
						UserID:    uuid.New(),
						ExpiresAt: time.Now().Add(time.Hour),
					}, nil
				}
				userRepo.GetUserByIDFn = func(ctx context.Context, id string) (db.User, error) {
					return db.User{
						ID:    uuid.New(),
						Email: "test@example.com",
					}, nil
				}
				tokenRepo.MarkAsUsedFn = func(ctx context.Context, tokenID uuid.UUID) error {
					return nil
				}
				userRepo.UpdatePasswordFn = func(ctx context.Context, id, passwordHash string) error {
					return errors.New("password update failed")
				}
			},
			expectedErr: errors.New("password update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userRepo := &MockPRUserRepository{}
			tokenRepo := &MockPRPasswordResetTokenRepository{}
			refreshTokenRepo := &MockPRRefreshTokenRepository{}

			if tt.setup != nil {
				tt.setup(userRepo, tokenRepo, refreshTokenRepo)
			}

			svc := service.NewPasswordResetService(
				userRepo,
				tokenRepo,
				refreshTokenRepo,
				&MockPREmailClient{},
				"https://example.com",
			)

			// Act
			err := svc.ConfirmReset(context.Background(), "valid-token", "newpassword123")

			// Assert
			if tt.expectedErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedErr, repository.ErrTokenNotFound) ||
					errors.Is(tt.expectedErr, repository.ErrTokenExpired) ||
					errors.Is(tt.expectedErr, repository.ErrTokenUsed) {
					assert.True(t, errors.Is(err, tt.expectedErr),
						"expected error %v, got %v", tt.expectedErr, err)
				} else {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}