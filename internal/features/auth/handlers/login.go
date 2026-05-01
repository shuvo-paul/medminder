package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

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

func LoginHandler(svc service.AuthService) func(context.Context, *LoginInput) (*LoginOutput, error) {
	return func(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
		if err := ValidateEmail(input.Email); err != nil {
			return nil, ErrInvalidEmail
		}

		result, err := svc.Login(ctx, input.Email, input.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				return nil, huma.Error401Unauthorized("Invalid email or password", err)
			}
			return nil, err
		}

		resp := &LoginOutput{}
		resp.AccessToken = result.AccessToken
		resp.RefreshToken = result.RefreshToken
		resp.User.ID = result.User.ID
		resp.User.Email = result.User.Email
		resp.User.DisplayName = result.User.DisplayName
		resp.User.EmailVerified = result.User.EmailVerified

		return resp, nil
	}
}