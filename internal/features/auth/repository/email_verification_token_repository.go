package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

// EmailVerificationTokenRepository defines the interface for email verification token data access.
type EmailVerificationTokenRepository interface {
	CreateToken(ctx context.Context, id uuid.UUID, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.EmailVerificationToken, error)
	FindValidToken(ctx context.Context, tokenHash string) (db.EmailVerificationToken, error)
	DeleteToken(ctx context.Context, id uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
	CountTokensCreatedToday(ctx context.Context, userID uuid.UUID) (int64, error)
}
