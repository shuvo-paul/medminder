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

	"github.com/google/uuid"

	emailqueue "github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/common/log"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	emailChangeTokenExpiry   = 24 * time.Hour
	emailChangeTokenByteSize = 32
)

type VerifyEmailChangeResult struct {
	AccessToken string
	User        struct {
		ID            uuid.UUID
		Email         string
		DisplayName   string
		EmailVerified bool
	}
}

type EmailChangeService interface {
	RequestEmailChange(ctx context.Context, userID uuid.UUID, newEmail, currentPassword string) error
	VerifyEmailChange(ctx context.Context, token string) (*VerifyEmailChangeResult, error)
	CancelEmailChange(ctx context.Context, userID uuid.UUID) error
}

type emailChangeService struct {
	userRepo    repository.UserRepository
	changeRepo  repository.EmailChangeRepository
	tokenSvc    TokenServiceInterface
	frontendURL string
}

var _ EmailChangeService = (*emailChangeService)(nil)

func NewEmailChangeService(
	userRepo repository.UserRepository,
	changeRepo repository.EmailChangeRepository,
	tokenSvc TokenServiceInterface,
	frontendURL string,
) *emailChangeService {
	return &emailChangeService{
		userRepo:    userRepo,
		changeRepo:  changeRepo,
		tokenSvc:    tokenSvc,
		frontendURL: frontendURL,
	}
}

func (s *emailChangeService) RequestEmailChange(ctx context.Context, userID uuid.UUID, newEmail, currentPassword string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("getting user: %w", err)
	}

	existingUser, err := s.userRepo.GetUserByEmail(ctx, newEmail)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("checking email existence: %w", err)
		}
	} else if existingUser.ID != user.ID {
		return ErrEmailExists
	}

	if !user.PasswordHash.Valid {
		return ErrNoPasswordSet
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(currentPassword)); err != nil {
		return ErrWrongPassword
	}

	pendingReq, err := s.changeRepo.GetPendingByUserID(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrTokenNotFound) {
		return fmt.Errorf("checking pending request: %w", err)
	}
	if err == nil {
		if err := s.changeRepo.Delete(ctx, pendingReq.ID); err != nil {
			log.Warn("failed to delete old email change request", log.F("error", err.Error()))
		}
	}

	tokenBytes := make([]byte, emailChangeTokenByteSize)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generating token: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	tokenHash, err := hashEmailChangeToken(rawToken)
	if err != nil {
		return fmt.Errorf("hashing token: %w", err)
	}

	expiresAt := time.Now().Add(emailChangeTokenExpiry)

	_, err = s.changeRepo.Create(ctx, userID, newEmail, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("storing email change request: %w", err)
	}

	verificationLink := fmt.Sprintf("%s/auth/verify-updated?token=%s", s.frontendURL, rawToken)
	emailBody := fmt.Sprintf("<p>You requested to change your email address to this email. Click <a href=\"%s\">here</a> to confirm this change. This link expires in 24 hours.</p>", verificationLink)

	emailqueue.QueueEmail(ctx, newEmail, "MedMinder Email Change Verification", emailBody)

	return nil
}

func (s *emailChangeService) VerifyEmailChange(ctx context.Context, token string) (*VerifyEmailChangeResult, error) {
	tokenHash, err := hashEmailChangeToken(token)
	if err != nil {
		return nil, fmt.Errorf("hashing token: %w", err)
	}

	changeReq, err := s.changeRepo.FindValidByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, changeReq.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	if err := s.userRepo.UpdateUserEmail(ctx, user.ID, changeReq.NewEmail); err != nil {
		return nil, fmt.Errorf("updating user email: %w", err)
	}

	if err := s.changeRepo.DeleteAllForUser(ctx, user.ID); err != nil {
		log.Warn("failed to delete email change requests", log.F("user_id", user.ID.String()), log.F("error", err.Error()))
	}

	accessToken, err := s.tokenSvc.GenerateAccessToken(user.ID, changeReq.NewEmail)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	result := &VerifyEmailChangeResult{}
	result.AccessToken = accessToken
	result.User.ID = user.ID
	result.User.Email = changeReq.NewEmail
	result.User.DisplayName = user.DisplayName
	result.User.EmailVerified = true

	return result, nil
}

func (s *emailChangeService) CancelEmailChange(ctx context.Context, userID uuid.UUID) error {
	pendingReq, err := s.changeRepo.GetPendingByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil
		}
		return fmt.Errorf("finding pending request: %w", err)
	}

	if err := s.changeRepo.Delete(ctx, pendingReq.ID); err != nil {
		return fmt.Errorf("deleting email change request: %w", err)
	}

	return nil
}

func hashEmailChangeToken(token string) (string, error) {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}
