package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

func RequestEmailChangeHandler(svc service.EmailChangeService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.RequestEmailChangeInput) (*dto.RequestEmailChangeOutput, error) {
	return func(ctx context.Context, input *dto.RequestEmailChangeInput) (*dto.RequestEmailChangeOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired token", nil)
		}

		if input.Body.NewEmail == "" {
			return nil, huma.Error400BadRequest("New email is required", nil)
		}
		if input.Body.CurrentPassword == "" {
			return nil, huma.Error400BadRequest("Current password is required", nil)
		}

		err = svc.RequestEmailChange(ctx, userID, input.Body.NewEmail, input.Body.CurrentPassword)
		if err != nil {
			if errors.Is(err, service.ErrEmailExists) {
				return nil, huma.Error409Conflict("Email already in use", err)
			}
			if errors.Is(err, service.ErrWrongPassword) {
				return nil, huma.Error400BadRequest("Current password is incorrect", err)
			}
			if errors.Is(err, service.ErrNoPasswordSet) {
				return nil, huma.Error400BadRequest("Cannot change email: no password set", err)
			}
			return nil, err
		}

		return &dto.RequestEmailChangeOutput{
			Body: struct {
				Message string `json:"message"`
			}{Message: "Verification email sent to your new email address"},
		}, nil
	}
}

func CancelEmailChangeHandler(svc service.EmailChangeService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.CancelEmailChangeInput) (*dto.CancelEmailChangeOutput, error) {
	return func(ctx context.Context, input *dto.CancelEmailChangeInput) (*dto.CancelEmailChangeOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired token", nil)
		}

		if err := svc.CancelEmailChange(ctx, userID); err != nil {
			return nil, err
		}

		return &dto.CancelEmailChangeOutput{
			Body: struct {
				Message string `json:"message"`
			}{Message: "Email change request cancelled"},
		}, nil
	}
}

func VerifyUpdatedEmailHandler(svc service.EmailChangeService) func(context.Context, *dto.VerifyEmailInput) (*dto.VerifyEmailOutput, error) {
	return func(ctx context.Context, input *dto.VerifyEmailInput) (*dto.VerifyEmailOutput, error) {
		if input.Body.Token == "" {
			return nil, huma.Error400BadRequest("Token is required", nil)
		}

		result, err := svc.VerifyEmailChange(ctx, input.Body.Token)
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

func GetPendingEmailChangeHandler(svc service.EmailChangeService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.GetPendingEmailChangeInput) (*dto.GetPendingEmailChangeOutput, error) {
	return func(ctx context.Context, input *dto.GetPendingEmailChangeInput) (*dto.GetPendingEmailChangeOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired token", nil)
		}

		newEmail, expiresAt, err := svc.GetPendingEmailChange(ctx, userID)
		if err != nil {
			if errors.Is(err, repository.ErrTokenNotFound) || errors.Is(err, repository.ErrTokenExpired) {
				return nil, huma.Error404NotFound("No pending email change request", err)
			}
			return nil, err
		}

		return &dto.GetPendingEmailChangeOutput{
			Body: struct {
				NewEmail  string `json:"new_email"`
				ExpiresAt string `json:"expires_at"`
			}{NewEmail: newEmail, ExpiresAt: expiresAt.Format(time.RFC3339)},
		}, nil
	}
}
