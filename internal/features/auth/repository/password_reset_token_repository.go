package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenUsed     = errors.New("token already used")
)

type PasswordResetTokenRepository interface {
	CreateToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.PasswordResetToken, error)
	FindValidToken(ctx context.Context, tokenHash string) (db.PasswordResetToken, error)
	MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type passwordResetTokenRepository struct {
	queries *db.Queries
}

func NewPasswordResetTokenRepository(queries *db.Queries) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{queries: queries}
}

func (r *passwordResetTokenRepository) CreateToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (db.PasswordResetToken, error) {
	return r.queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
}

func (r *passwordResetTokenRepository) FindValidToken(ctx context.Context, tokenHash string) (db.PasswordResetToken, error) {
	token, err := r.queries.FindPasswordResetTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.PasswordResetToken{}, ErrTokenNotFound
		}
		return db.PasswordResetToken{}, err
	}
	if time.Now().After(token.ExpiresAt) {
		return db.PasswordResetToken{}, ErrTokenExpired
	}
	if token.UsedAt.Valid {
		return db.PasswordResetToken{}, ErrTokenUsed
	}
	return token, nil
}

func (r *passwordResetTokenRepository) MarkAsUsed(ctx context.Context, tokenID uuid.UUID) error {
	return r.queries.MarkPasswordResetTokenUsed(ctx, tokenID)
}

func (r *passwordResetTokenRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeletePasswordResetTokensByUserID(ctx, userID)
}
