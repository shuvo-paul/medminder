package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

type PasswordResetDeps struct {
	Svc service.PasswordResetService
}

func NewPasswordResetDeps(svc service.PasswordResetService) PasswordResetDeps {
	return PasswordResetDeps{Svc: svc}
}

func RegisterPasswordResetRoutes(api huma.API, deps PasswordResetDeps) {
	huma.Register(api, huma.Operation{
		OperationID: "request-password-reset",
		Method:      http.MethodPost,
		Path:        "/api/auth/password/reset/request",
		Summary:     "Request password reset",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *dto.PasswordResetRequestInput) (*dto.PasswordResetRequestOutput, error) {
		if err := deps.Svc.RequestReset(ctx, input.Body.Email); err != nil {
			return nil, err
		}
		return &dto.PasswordResetRequestOutput{Body: struct {
			Message string `json:"message"`
		}{Message: "If the email exists, a reset link has been sent"}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "confirm-password-reset",
		Method:      http.MethodPost,
		Path:        "/api/auth/password/reset/confirm",
		Summary:     "Confirm password reset",
		Tags:        []string{"auth"},
	}, func(ctx context.Context, input *dto.PasswordResetConfirmInput) (*dto.PasswordResetConfirmOutput, error) {
		if err := ValidatePassword(input.Body.NewPassword); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, err.Error())
		}

		if err := deps.Svc.ConfirmReset(ctx, input.Body.Token, input.Body.NewPassword); err != nil {
			if errors.Is(err, repository.ErrTokenNotFound) ||
				errors.Is(err, repository.ErrTokenExpired) ||
				errors.Is(err, repository.ErrTokenUsed) {
				return nil, huma.NewError(http.StatusBadRequest, "Invalid or expired token")
			}
			return nil, err
		}

		return &dto.PasswordResetConfirmOutput{Body: struct {
			Message string `json:"message"`
		}{Message: "Password has been reset successfully"}}, nil
	})
}
