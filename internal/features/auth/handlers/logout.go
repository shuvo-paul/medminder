package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/service"
)

// LogoutHandler returns a thin HTTP adapter that delegates logout to AuthService.
func LogoutHandler(svc service.AuthService) func(context.Context, uuid.UUID) error {
	return func(ctx context.Context, userID uuid.UUID) error {
		if userID == uuid.Nil {
			return huma.Error401Unauthorized("Invalid user ID", nil)
		}

		if err := svc.Logout(ctx, userID); err != nil {
			if errors.Is(err, service.ErrLogoutFailed) {
				return huma.Error500InternalServerError("Failed to logout", err)
			}
			return err
		}

		return nil
	}
}
