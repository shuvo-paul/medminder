package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type EmailChangeRepository interface {
	Create(ctx context.Context, userID uuid.UUID, newEmail, tokenHash string, expiresAt time.Time) (db.EmailChangeRequest, error)
	FindValidByTokenHash(ctx context.Context, tokenHash string) (db.EmailChangeRequest, error)
	GetPendingByUserID(ctx context.Context, userID uuid.UUID) (db.EmailChangeRequest, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type emailChangeRepository struct {
	queries *db.Queries
}

func NewEmailChangeRepository(queries *db.Queries) EmailChangeRepository {
	return &emailChangeRepository{queries: queries}
}

func (r *emailChangeRepository) Create(ctx context.Context, userID uuid.UUID, newEmail, tokenHash string, expiresAt time.Time) (db.EmailChangeRequest, error) {
	return r.queries.CreateEmailChangeRequest(ctx, db.CreateEmailChangeRequestParams{
		UserID:    userID,
		NewEmail:  newEmail,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
}

func (r *emailChangeRepository) FindValidByTokenHash(ctx context.Context, tokenHash string) (db.EmailChangeRequest, error) {
	token, err := r.queries.FindValidEmailChangeRequestByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.EmailChangeRequest{}, ErrTokenNotFound
		}
		return db.EmailChangeRequest{}, err
	}
	if time.Now().After(token.ExpiresAt) {
		return db.EmailChangeRequest{}, ErrTokenExpired
	}
	return token, nil
}

func (r *emailChangeRepository) GetPendingByUserID(ctx context.Context, userID uuid.UUID) (db.EmailChangeRequest, error) {
	token, err := r.queries.GetPendingEmailChangeRequestByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.EmailChangeRequest{}, ErrTokenNotFound
		}
		return db.EmailChangeRequest{}, err
	}
	if time.Now().After(token.ExpiresAt) {
		return db.EmailChangeRequest{}, ErrTokenExpired
	}
	return token, nil
}

func (r *emailChangeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteEmailChangeRequest(ctx, id)
}

func (r *emailChangeRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteEmailChangeRequestsByUserID(ctx, userID)
}
