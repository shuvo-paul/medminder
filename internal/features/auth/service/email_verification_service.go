package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	emailqueue "github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/common/log"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
)

const (
	emailVerificationTokenExpiry = 24 * time.Hour
	maxVerificationsPerDay      = 3
	verificationTokenByteSize   = 32
)

type VerifyResult struct {
	AccessToken string
	User        struct {
		ID            uuid.UUID
		Email         string
		DisplayName   string
		EmailVerified bool
	}
}

// EmailVerificationService defines the interface for email verification operations.
type EmailVerificationService interface {
	VerifyEmail(ctx context.Context, token string) (*VerifyResult, error)
	ResendVerification(ctx context.Context, userID uuid.UUID, email string) error
}

// emailVerificationService implements EmailVerificationService.
type emailVerificationService struct {
	userRepo         repository.UserRepository
	tokenRepo        repository.EmailVerificationTokenRepository
	tokenSvc         TokenServiceInterface
	frontendURL      string
}

// NewEmailVerificationService creates a new EmailVerificationService.
func NewEmailVerificationService(
	userRepo repository.UserRepository,
	tokenRepo repository.EmailVerificationTokenRepository,
	tokenSvc TokenServiceInterface,
	frontendURL string,
) *emailVerificationService {
	return &emailVerificationService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		tokenSvc:    tokenSvc,
		frontendURL: frontendURL,
	}
}

// VerifyEmail validates a verification token and marks the user's email as verified.
// It returns a new access token upon successful verification.
func (s *emailVerificationService) VerifyEmail(ctx context.Context, token string) (*VerifyResult, error) {
	tokenHash, err := hashEmailVerificationToken(token)
	if err != nil {
		return nil, fmt.Errorf("hashing token: %w", err)
	}

	storedToken, err := s.tokenRepo.FindValidToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) || errors.Is(err, repository.ErrTokenExpired) {
			return nil, err
		}
		return nil, fmt.Errorf("finding valid token: %w", err)
	}

	user, err := s.userRepo.GetUserByID(ctx, storedToken.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	if err := s.userRepo.VerifyEmail(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("verifying email: %w", err)
	}

	if err := s.tokenRepo.DeleteToken(ctx, storedToken.ID); err != nil {
		log.Warn("failed to delete used token", log.F("token_id", storedToken.ID.String()), log.F("error", err.Error()))
	}
	if err := s.tokenRepo.DeleteAllForUser(ctx, user.ID); err != nil {
		log.Warn("failed to delete user tokens", log.F("user_id", user.ID.String()), log.F("error", err.Error()))
	}

	accessToken, err := s.tokenSvc.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	result := &VerifyResult{}
	result.AccessToken = accessToken
	result.User.ID = user.ID
	result.User.Email = user.Email
	result.User.DisplayName = user.DisplayName
	result.User.EmailVerified = true

	return result, nil
}

// ResendVerification sends a new verification email if the daily limit hasn't been exceeded.
func (s *emailVerificationService) ResendVerification(ctx context.Context, userID uuid.UUID, email string) error {
	count, err := s.tokenRepo.CountTokensCreatedToday(ctx, userID)
	if err != nil {
		return fmt.Errorf("checking token count: %w", err)
	}

	if count >= maxVerificationsPerDay {
		return ErrRateLimitExceeded
	}

	tokenBytes := make([]byte, verificationTokenByteSize)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generating token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	tokenHash, err := hashEmailVerificationToken(rawToken)
	if err != nil {
		return fmt.Errorf("hashing token: %w", err)
	}

	expiresAt := time.Now().Add(emailVerificationTokenExpiry)

	_, err = s.tokenRepo.CreateToken(ctx, uuid.New(), userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("storing token: %w", err)
	}

	verificationLink := fmt.Sprintf("%s/auth/verify-email?token=%s", s.frontendURL, rawToken)
	emailBody := fmt.Sprintf("<p>Click <a href=\"%s\">here</a> to verify your email address.</p>", verificationLink)

	emailqueue.QueueEmail(ctx, email, "MedMinder Email Verification", emailBody)

	return nil
}

// hashEmailVerificationToken hashes a token using SHA256.
func hashEmailVerificationToken(token string) (string, error) {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}