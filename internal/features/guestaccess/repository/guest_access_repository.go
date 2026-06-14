package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type GuestAccessRepository interface {
	Create(ctx context.Context, arg db.CreateGuestAccessTokenParams) (db.GuestAccessToken, error)
	GetByHash(ctx context.Context, tokenHash string) (db.GuestAccessToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (db.GuestAccessToken, error)
	ListByProfile(ctx context.Context, profileID uuid.UUID) ([]db.GuestAccessToken, error)
	UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type guestAccessRepository struct {
	queries *db.Queries
}

func NewGuestAccessRepository(queries *db.Queries) GuestAccessRepository {
	return &guestAccessRepository{queries: queries}
}

func (r *guestAccessRepository) Create(ctx context.Context, arg db.CreateGuestAccessTokenParams) (db.GuestAccessToken, error) {
	return r.queries.CreateGuestAccessToken(ctx, arg)
}

func (r *guestAccessRepository) GetByHash(ctx context.Context, tokenHash string) (db.GuestAccessToken, error) {
	return r.queries.GetGuestAccessTokenByHash(ctx, tokenHash)
}

func (r *guestAccessRepository) GetByID(ctx context.Context, id uuid.UUID) (db.GuestAccessToken, error) {
	return r.queries.GetGuestAccessTokenByID(ctx, id)
}

func (r *guestAccessRepository) ListByProfile(ctx context.Context, profileID uuid.UUID) ([]db.GuestAccessToken, error) {
	return r.queries.ListGuestAccessTokensByProfile(ctx, profileID)
}

func (r *guestAccessRepository) UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateGuestAccessTokenLastUsedAt(ctx, id)
}

func (r *guestAccessRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteGuestAccessToken(ctx, id)
}

func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func NewNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

func NewPermissions(perms []string) (json.RawMessage, error) {
	return json.Marshal(perms)
}
