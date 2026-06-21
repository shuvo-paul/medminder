package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/common/auth"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

func ResendVerificationHandler(svc service.EmailVerificationService, tokenSvc auth.TokenValidator) func(context.Context, *dto.ResendVerificationInput) (*dto.ResendVerificationOutput, error) {
	return func(ctx context.Context, input *dto.ResendVerificationInput) (*dto.ResendVerificationOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		email, err := auth.ExtractEmail(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
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
