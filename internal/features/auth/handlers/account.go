package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	"github.com/shuvo-paul/medminder/internal/features/auth/dto"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
	profilesvc "github.com/shuvo-paul/medminder/internal/features/profiles/service"
)

func DeleteAccountHandler(authSvc service.AuthService, profileSvc profilesvc.ProfileService, tokenSvc service.TokenServiceInterface) func(context.Context, *dto.DeleteAccountInput) (*dto.DeleteAccountOutput, error) {
	return func(ctx context.Context, input *dto.DeleteAccountInput) (*dto.DeleteAccountOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		if err := authSvc.VerifyPassword(ctx, userID, input.Body.Password); err != nil {
			if errors.Is(err, service.ErrWrongPassword) {
				return nil, huma.Error403Forbidden("current password is incorrect", err)
			}
			if errors.Is(err, service.ErrUserNotFound) {
				return nil, huma.Error404NotFound("user not found", err)
			}
			return nil, huma.Error500InternalServerError("failed to verify password", err)
		}

		if err := profileSvc.HandleAccountDeletion(ctx, userID); err != nil {
			return nil, huma.Error500InternalServerError("failed to handle profile cleanup", err)
		}

		if err := authSvc.DeleteAccount(ctx, userID); err != nil {
			return nil, huma.Error500InternalServerError("failed to delete account", err)
		}

		return &dto.DeleteAccountOutput{
			Body: struct {
				Message string `json:"message"`
			}{Message: "account deleted successfully"},
		}, nil
	}
}
