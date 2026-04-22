package handlers

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	db "github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type LoginHandlerRepo interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error)
}

type LoginInput struct {
	Email    string `json:"email" minLength:"1" maxLength:"255"`
	Password string `json:"password" minLength:"1"`
}

type LoginOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID            uuid.UUID `json:"id"`
		Email         string    `json:"email"`
		DisplayName   string    `json:"display_name"`
		EmailVerified bool      `json:"email_verified"`
	} `json:"user"`
}

func LoginHandler(userRepo repository.UserRepository, tokenRepo repository.RefreshTokenRepository, tokenSvc service.TokenServiceInterface) func(context.Context, *LoginInput) (*LoginOutput, error) {
	return func(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
		if err := ValidateEmail(input.Email); err != nil {
			return nil, ErrInvalidEmail
		}

		user, err := userRepo.GetUserByEmail(ctx, input.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrInvalidCredentials
			}
			return nil, err
		}

		if !user.PasswordHash.Valid || user.PasswordHash.String == "" {
			return nil, ErrInvalidCredentials
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(input.Password)); err != nil {
			return nil, ErrInvalidCredentials
		}

		accessToken, err := tokenSvc.GenerateAccessToken(user.ID, user.Email)
		if err != nil {
			return nil, err
		}

		refreshToken, err := tokenSvc.GenerateRefreshToken()
		if err != nil {
			return nil, err
		}

		tokenHash := tokenSvc.HashRefreshToken(refreshToken)
		expiresAt := time.Now().Add(RefreshTokenExpiry)
		if _, err := tokenRepo.CreateRefreshToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
			return nil, err
		}

		resp := &LoginOutput{}
		resp.AccessToken = accessToken
		resp.RefreshToken = refreshToken
		resp.User.ID = user.ID
		resp.User.Email = user.Email
		resp.User.DisplayName = user.DisplayName
		resp.User.EmailVerified = user.EmailVerified.Bool

		return resp, nil
	}
}
