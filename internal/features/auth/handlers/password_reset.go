package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/crypto/bcrypt"

	emailclient "github.com/shuvo-paul/medminder/internal/common/email"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
)

type PasswordResetDeps struct {
	UserRepo         repository.UserRepository
	TokenRepo        repository.PasswordResetTokenRepository
	RefreshTokenRepo repository.RefreshTokenRepository
	EmailClient      emailclient.EmailClient
	FrontendURL      string
}

func NewPasswordResetDeps(
	userRepo repository.UserRepository,
	tokenRepo repository.PasswordResetTokenRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	emailClient emailclient.EmailClient,
	frontendURL string,
) PasswordResetDeps {
	return PasswordResetDeps{
		UserRepo:         userRepo,
		TokenRepo:        tokenRepo,
		RefreshTokenRepo: refreshTokenRepo,
		EmailClient:      emailClient,
		FrontendURL:      frontendURL,
	}
}

func RegisterPasswordResetRoutes(api huma.API, deps PasswordResetDeps) {
	huma.Register(api, huma.Operation{
		OperationID: "request-password-reset",
		Method:      http.MethodPost,
		Path:        "/api/auth/password/reset/request",
		Summary:     "Request password reset",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *dto.PasswordResetRequestInput) (*dto.PasswordResetRequestOutput, error) {
		user, err := deps.UserRepo.GetUserByEmail(ctx, input.Body.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return &dto.PasswordResetRequestOutput{Body: struct {
					Message string `json:"message"`
				}{Message: "If the email exists, a reset link has been sent"}}, nil
			}
			return nil, err
		}

		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return nil, fmt.Errorf("generating token: %w", err)
		}
		token := hex.EncodeToString(tokenBytes)
		tokenHash, err := hashToken(token)
		if err != nil {
			return nil, fmt.Errorf("hashing token: %w", err)
		}

expiresAt := time.Now().Add(1 * time.Hour)

	err = deps.TokenRepo.DeleteAllForUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("cleaning up existing tokens: %w", err)
	}

	_, err = deps.TokenRepo.CreateToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("storing token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/auth/reset-password?token=%s", deps.FrontendURL, token)
	emailBody := fmt.Sprintf("<p>Click <a href=\"%s\">here</a> to reset your password. This link expires in 1 hour.</p>", resetLink)
	err = deps.EmailClient.SendEmail(ctx, user.Email, "MedMinder Password Reset", emailBody)
	if err != nil {
		log.Printf("failed to send password reset email, cleaning up token: %v", err)
		if cleanupErr := deps.TokenRepo.DeleteAllForUser(ctx, user.ID); cleanupErr != nil {
			log.Printf("failed to cleanup token after email failure: %v", cleanupErr)
		}
		return &dto.PasswordResetRequestOutput{Body: struct {
			Message string `json:"message"`
		}{Message: "If the email exists, a reset link has been sent"}}, nil
	}

		return &dto.PasswordResetRequestOutput{Body: struct {
			Message string `json:"message"`
		}{Message: "If the email exists, a reset link has been sent"}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "confirm-password-reset",
		Method:      http.MethodPost,
		Path:        "/api/auth/password/reset/confirm",
		Summary:     "Confirm password reset",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *dto.PasswordResetConfirmInput) (*dto.PasswordResetConfirmOutput, error) {
		if err := ValidatePassword(input.Body.NewPassword); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, err.Error())
		}

		tokenHash, err := hashToken(input.Body.Token)
		if err != nil {
			return nil, fmt.Errorf("processing token: %w", err)
		}

		storedToken, err := deps.TokenRepo.FindValidToken(ctx, tokenHash)
		if err != nil {
			if errors.Is(err, repository.ErrTokenNotFound) || errors.Is(err, repository.ErrTokenExpired) || errors.Is(err, repository.ErrTokenUsed) {
				return nil, huma.NewError(http.StatusBadRequest, "Invalid or expired token")
			}
			return nil, err
		}

		user, err := deps.UserRepo.GetUserByID(ctx, storedToken.UserID.String())
		if err != nil {
			return nil, err
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Body.NewPassword), BcryptCost)
		if err != nil {
			return nil, fmt.Errorf("hashing password: %w", err)
		}
		err = deps.TokenRepo.MarkAsUsed(ctx, storedToken.ID)
		if err != nil {
			return nil, fmt.Errorf("marking token as used: %w", err)
		}

		err = deps.UserRepo.UpdatePassword(ctx, user.ID.String(), string(passwordHash))
		if err != nil {
			return nil, fmt.Errorf("updating password: %w", err)
		}

		err = deps.RefreshTokenRepo.DeleteAllForUser(ctx, user.ID)
		if err != nil {
			log.Printf("failed to delete refresh tokens: %v", err)
		}

		return &dto.PasswordResetConfirmOutput{Body: struct {
			Message string `json:"message"`
		}{Message: "Password has been reset successfully"}}, nil
	})
}

func hashToken(token string) (string, error) {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}
