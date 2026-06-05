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
			return nil, huma.Error400BadRequest("invalid password", err)
		}

		if err := deps.Svc.ConfirmReset(ctx, input.Body.Token, input.Body.NewPassword); err != nil {
			if errors.Is(err, repository.ErrTokenNotFound) ||
				errors.Is(err, repository.ErrTokenExpired) ||
				errors.Is(err, repository.ErrTokenUsed) {
				return nil, huma.Error400BadRequest("Invalid or expired token", err)
			}
			return nil, err
		}

		return &dto.PasswordResetConfirmOutput{Body: struct {
			Message string `json:"message"`
		}{Message: "Password has been reset successfully"}}, nil
	})
}

func SetPasswordHandler(authSvc service.AuthService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.SetPasswordInput) (*dto.SetPasswordOutput, error) {
	return func(ctx context.Context, input *dto.SetPasswordInput) (*dto.SetPasswordOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		if err := ValidatePassword(input.Body.Password); err != nil {
			return nil, huma.Error400BadRequest("invalid password", err)
		}

		if err := authSvc.SetPassword(ctx, userID, input.Body.Password); err != nil {
			return nil, huma.Error500InternalServerError("failed to set password", err)
		}

		return &dto.SetPasswordOutput{
			Body: struct {
				Message string `json:"message" doc:"Success message"`
			}{Message: "password set successfully"},
		}, nil
	}
}

func ChangePasswordHandler(authSvc service.AuthService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.ChangePasswordInput) (*dto.ChangePasswordOutput, error) {
	return func(ctx context.Context, input *dto.ChangePasswordInput) (*dto.ChangePasswordOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		if err := ValidatePassword(input.Body.NewPassword); err != nil {
			return nil, huma.Error400BadRequest("invalid new password", err)
		}

		if input.Body.NewPassword != input.Body.ConfirmPassword {
			return nil, huma.Error400BadRequest("passwords do not match")
		}

		if err := authSvc.ChangePassword(ctx, userID, input.Body.CurrentPassword, input.Body.NewPassword); err != nil {
			if errors.Is(err, service.ErrNoPasswordSet) {
				return nil, huma.Error400BadRequest("no password set", err)
			}
			if errors.Is(err, service.ErrWrongPassword) {
				return nil, huma.Error403Forbidden("current password is incorrect", err)
			}
			if errors.Is(err, service.ErrUserNotFound) {
				return nil, huma.Error404NotFound("user not found", err)
			}
			if errors.Is(err, service.ErrSamePassword) {
				return nil, huma.Error400BadRequest("new password must differ from current password", err)
			}
			return nil, huma.Error500InternalServerError("failed to change password", err)
		}

		return &dto.ChangePasswordOutput{
			Body: struct {
				Message string `json:"message"`
			}{Message: "password changed successfully"},
		}, nil
	}
}
