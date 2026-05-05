package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

// LogoutHandler returns a thin HTTP adapter that extracts the user ID from
// the Authorization header and delegates logout to AuthService.
func LogoutHandler(authSvc service.AuthService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.LogoutInput) (*dto.LogoutOutput, error) {
	return func(ctx context.Context, input *dto.LogoutInput) (*dto.LogoutOutput, error) {
		authHeader := input.Authorization
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return nil, huma.Error401Unauthorized("Invalid authorization header", nil)
		}
		tokenString := authHeader[7:]

		claims, err := tokenSvc.ValidateAccessToken(tokenString)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		userIDStr, ok := claims["sub"].(string)
		if !ok {
			return nil, huma.Error401Unauthorized("Invalid access token", nil)
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid user ID in token", nil)
		}

		if userID == uuid.Nil {
			return nil, huma.Error401Unauthorized("Invalid user ID", nil)
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
