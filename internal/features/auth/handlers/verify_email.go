package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

func VerifyEmailHandler(svc service.EmailVerificationService) func(context.Context, *dto.VerifyEmailInput) (*dto.VerifyEmailOutput, error) {
	return func(ctx context.Context, input *dto.VerifyEmailInput) (*dto.VerifyEmailOutput, error) {
		if input.Body.Token == "" {
			return nil, huma.Error400BadRequest("Token is required", nil)
		}

		result, err := svc.VerifyEmail(ctx, input.Body.Token)
		if err != nil {
			if errors.Is(err, repository.ErrTokenNotFound) {
				return nil, huma.Error400BadRequest("Invalid or expired token", err)
			}
			if errors.Is(err, repository.ErrTokenExpired) {
				return nil, huma.Error400BadRequest("Invalid or expired token", err)
			}
			return nil, err
		}

		return &dto.VerifyEmailOutput{
			Body: struct {
				AccessToken string `json:"access_token"`
			}{AccessToken: result.AccessToken},
		}, nil
	}
}