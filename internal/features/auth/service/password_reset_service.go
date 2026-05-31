package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	emailclient "github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/common/log"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
)

const (
	tokenExpiry = 1 * time.Hour
)

// PasswordResetService defines the interface for password reset operations.
type PasswordResetService interface {
	RequestReset(ctx context.Context, email string) error
	ConfirmReset(ctx context.Context, token, newPassword string) error
}

// passwordResetService implements PasswordResetService.
type passwordResetService struct {
	userRepo         repository.UserRepository
	tokenRepo        repository.PasswordResetTokenRepository
	refreshTokenRepo repository.RefreshTokenRepository
	emailClient      emailclient.EmailClient
	frontendURL      string
}

var _ PasswordResetService = (*passwordResetService)(nil)

// NewPasswordResetService creates a new PasswordResetService.
func NewPasswordResetService(
	userRepo repository.UserRepository,
	tokenRepo repository.PasswordResetTokenRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	emailClient emailclient.EmailClient,
	frontendURL string,
) *passwordResetService {
	return &passwordResetService{
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		refreshTokenRepo: refreshTokenRepo,
		emailClient:      emailClient,
		frontendURL:      frontendURL,
	}
}

// RequestReset initiates a password reset request for the given email.
// It generates a token, stores it, and sends a reset email.
// Returns nil even if the email is not found to prevent email enumeration.
func (s *passwordResetService) RequestReset(ctx context.Context, email string) error {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return nil to prevent email enumeration
			return nil
		}
		return fmt.Errorf("getting user by email: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generating token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	tokenHash, err := hashToken(token)
	if err != nil {
		return fmt.Errorf("hashing token: %w", err)
	}

	expiresAt := time.Now().Add(tokenExpiry)

	if err := s.tokenRepo.DeleteAllForUser(ctx, user.ID); err != nil {
		return fmt.Errorf("cleaning up existing tokens: %w", err)
	}

	_, err = s.tokenRepo.CreateToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("storing token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/auth/reset-password?token=%s", s.frontendURL, token)
	emailBody := fmt.Sprintf("<p>Click <a href=\"%s\">here</a> to reset your password. This link expires in 1 hour.</p>", resetLink)
	if err := s.emailClient.SendEmail(ctx, user.Email, "MedMinder Password Reset", emailBody); err != nil {
		log.Warn("failed to send password reset email, cleaning up token", log.F("error", err.Error()))
		if cleanupErr := s.tokenRepo.DeleteAllForUser(ctx, user.ID); cleanupErr != nil {
			log.Warn("failed to cleanup token after email failure", log.F("error", cleanupErr.Error()))
		}
		// Return nil to prevent email enumeration - user sees success message anyway
		return nil
	}

	return nil
}

// ConfirmReset confirms a password reset with the given token and new password.
// It validates the token, updates the user's password, and invalidates all refresh tokens.
func (s *passwordResetService) ConfirmReset(ctx context.Context, token, newPassword string) error {
	tokenHash, err := hashToken(token)
	if err != nil {
		return fmt.Errorf("processing token: %w", err)
	}

	storedToken, err := s.tokenRepo.FindValidToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) ||
			errors.Is(err, repository.ErrTokenExpired) ||
			errors.Is(err, repository.ErrTokenUsed) {
			return err
		}
		return fmt.Errorf("finding valid token: %w", err)
	}

	user, err := s.userRepo.GetUserByID(ctx, storedToken.UserID.String())
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.tokenRepo.MarkAsUsed(ctx, storedToken.ID); err != nil {
		return fmt.Errorf("marking token as used: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID.String(), string(passwordHash)); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	if err := s.refreshTokenRepo.DeleteUserRefreshTokens(ctx, user.ID); err != nil {
		log.Warn("failed to delete refresh tokens", log.F("error", err.Error()))
	}

	return nil
}

// hashToken hashes a token using SHA256.
func hashToken(token string) (string, error) {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}
