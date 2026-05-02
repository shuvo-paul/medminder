package repository

import (
	"context"
	"database/sql"
	"errors"
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

type emailVerificationTokenRepository struct {
	queries *db.Queries
}

func NewEmailVerificationTokenRepository(queries *db.Queries) EmailVerificationTokenRepository {
	return &emailVerificationTokenRepository{queries: queries}
}

func (r *emailVerificationTokenRepository) CreateToken(ctx context.Context, id uuid.UUID, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.EmailVerificationToken, error) {
	return r.queries.CreateEmailVerificationToken(ctx, db.CreateEmailVerificationTokenParams{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
}

func (r *emailVerificationTokenRepository) FindValidToken(ctx context.Context, tokenHash string) (db.EmailVerificationToken, error) {
	token, err := r.queries.FindValidEmailVerificationToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.EmailVerificationToken{}, ErrTokenNotFound
		}
		return db.EmailVerificationToken{}, err
	}
	if time.Now().After(token.ExpiresAt) {
		return db.EmailVerificationToken{}, ErrTokenExpired
	}
	return token, nil
}

func (r *emailVerificationTokenRepository) DeleteToken(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteEmailVerificationToken(ctx, id)
}

func (r *emailVerificationTokenRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteAllEmailVerificationTokensForUser(ctx, userID)
}

func (r *emailVerificationTokenRepository) CountTokensCreatedToday(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.CountEmailVerificationTokensCreatedToday(ctx, userID)
}
