package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
)

var ErrLogoutFailed = errors.New("logout failed")

type LogoutInput struct {
	UserID uuid.UUID
}

func LogoutHandler(repo repository.RefreshTokenRepository) func(context.Context, *LogoutInput) error {
	return func(ctx context.Context, input *LogoutInput) error {
		if input.UserID == uuid.Nil {
			return ErrLogoutFailed
		}

		if err := repo.DeleteUserRefreshTokens(ctx, input.UserID); err != nil {
			return errors.Join(ErrLogoutFailed, err)
		}

		return nil
	}
}
