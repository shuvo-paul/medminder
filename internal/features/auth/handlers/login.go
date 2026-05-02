package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

func LoginHandler(svc service.AuthService) func(context.Context, *dto.LoginInput) (*dto.LoginOutput, error) {
	return func(ctx context.Context, input *dto.LoginInput) (*dto.LoginOutput, error) {
		if err := ValidateEmail(input.Body.Email); err != nil {
			return nil, huma.Error400BadRequest("Invalid email format", err)
		}

		result, err := svc.Login(ctx, input.Body.Email, input.Body.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				return nil, huma.Error401Unauthorized("Invalid email or password", err)
			}
			return nil, err
		}

		resp := &dto.LoginOutput{}
		resp.Body.AccessToken = result.AccessToken
		resp.Body.RefreshToken = result.RefreshToken
		resp.Body.User.ID = result.User.ID
		resp.Body.User.Email = result.User.Email
		resp.Body.User.DisplayName = result.User.DisplayName
		resp.Body.User.EmailVerified = result.User.EmailVerified

		return resp, nil
	}
}
