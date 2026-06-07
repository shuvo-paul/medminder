package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type ProfileRepository interface {
	CreateProfile(ctx context.Context, ownerUserID uuid.UUID, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error)
	GetProfileByID(ctx context.Context, id uuid.UUID) (db.Profile, error)
	ListProfilesByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]db.Profile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error)
	DeleteProfile(ctx context.Context, id uuid.UUID) error
}

type profileRepository struct {
	queries *db.Queries
}

func NewProfileRepository(queries *db.Queries) ProfileRepository {
	return &profileRepository{queries: queries}
}

func (r *profileRepository) CreateProfile(ctx context.Context, ownerUserID uuid.UUID, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error) {
	return r.queries.CreateProfile(ctx, db.CreateProfileParams{
		OwnerUserID: ownerUserID,
		Name:        name,
		DateOfBirth: dateOfBirth,
		Timezone:    timezone,
	})
}

func (r *profileRepository) GetProfileByID(ctx context.Context, id uuid.UUID) (db.Profile, error) {
	return r.queries.GetProfileByID(ctx, id)
}

func (r *profileRepository) ListProfilesByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]db.Profile, error) {
	return r.queries.ListProfilesByOwner(ctx, ownerUserID)
}

func (r *profileRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error) {
	return r.queries.UpdateProfile(ctx, db.UpdateProfileParams{
		ID:          id,
		Name:        name,
		DateOfBirth: dateOfBirth,
		Timezone:    timezone,
	})
}

func (r *profileRepository) DeleteProfile(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteProfile(ctx, id)
}
