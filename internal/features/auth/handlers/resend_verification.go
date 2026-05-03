package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

func ResendVerificationHandler(svc service.EmailVerificationService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.ResendVerificationInput) (*dto.ResendVerificationOutput, error) {
	return func(ctx context.Context, input *dto.ResendVerificationInput) (*dto.ResendVerificationOutput, error) {
		authHeader := input.Body.Authorization
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

		email, ok := claims["email"].(string)
		if !ok {
			return nil, huma.Error401Unauthorized("Invalid email in token", nil)
		}

		if err := svc.ResendVerification(ctx, userID, email); err != nil {
			if errors.Is(err, service.ErrRateLimitExceeded) {
				return nil, huma.Error429TooManyRequests("Daily verification limit exceeded", err)
			}
			return nil, err
		}

		return &dto.ResendVerificationOutput{
			Body: struct {
				Message string `json:"message"`
			}{Message: "Verification email sent"},
		}, nil
	}
}