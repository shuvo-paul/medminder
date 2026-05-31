package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id uuid.UUID) error
	DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
}

type refreshTokenRepository struct {
	queries *db.Queries
}

func NewRefreshTokenRepository(queries *db.Queries) RefreshTokenRepository {
	return &refreshTokenRepository{queries: queries}
}

func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.CreateRefreshTokenRow, error) {
	return r.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
}

func (r *refreshTokenRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (db.RefreshToken, error) {
	return r.queries.GetRefreshTokenByHash(ctx, tokenHash)
}

func (r *refreshTokenRepository) DeleteRefreshToken(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteRefreshToken(ctx, id)
}

func (r *refreshTokenRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteUserRefreshTokens(ctx, userID)
}
