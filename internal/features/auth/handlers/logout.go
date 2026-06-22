package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/common/auth"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

// LogoutHandler returns a thin HTTP adapter that extracts the user ID from
// the Authorization header and delegates logout to AuthService.
func LogoutHandler(authSvc service.AuthService, tokenSvc auth.TokenValidator) func(context.Context, *dto.LogoutInput) (*dto.LogoutOutput, error) {
	return func(ctx context.Context, input *dto.LogoutInput) (*dto.LogoutOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		if err := authSvc.Logout(ctx, userID); err != nil {
			if errors.Is(err, service.ErrLogoutFailed) {
				return nil, huma.Error500InternalServerError("Failed to logout", err)
			}
			return nil, err
		}

		return &dto.LogoutOutput{Body: struct {
			Message string `json:"message"`
		}{Message: "logged out successfully"}}, nil
	}
}
