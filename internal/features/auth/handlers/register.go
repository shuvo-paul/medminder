package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

func RegisterHandler(svc service.AuthService) func(context.Context, *dto.RegisterInput) (*dto.RegisterOutput, error) {
	return func(ctx context.Context, input *dto.RegisterInput) (*dto.RegisterOutput, error) {
		if err := ValidateEmail(input.Body.Email); err != nil {
			return nil, huma.Error400BadRequest("Invalid email format", err)
		}
		if err := ValidatePassword(input.Body.Password); err != nil {
			return nil, huma.Error400BadRequest("Password must be at least 8 characters with 1 uppercase, 1 lowercase, and 1 number", err)
		}
		if err := ValidateDisplayName(input.Body.DisplayName); err != nil {
			return nil, huma.Error400BadRequest("Display name must be 1-100 characters", err)
		}

		result, err := svc.Register(ctx, input.Body.Email, input.Body.DisplayName, input.Body.Password)
		if err != nil {
			if errors.Is(err, service.ErrEmailExists) {
				return nil, huma.Error409Conflict("Email already exists", err)
			}
			return nil, err
		}

		resp := &dto.RegisterOutput{}
		resp.Body.User.ID = result.User.ID
		resp.Body.User.Email = result.User.Email
		resp.Body.User.DisplayName = result.User.DisplayName
		resp.Body.User.EmailVerified = result.User.EmailVerified

		return resp, nil
	}
}
