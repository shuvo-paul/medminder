package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type ProfileRepository interface {
	CreateProfile(ctx context.Context, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error)
	CreateProfilePermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permissions json.RawMessage) error
	CreateProfileWithPermission(ctx context.Context, name string, dateOfBirth sql.NullTime, timezone string, userID uuid.UUID, permissions json.RawMessage) (db.Profile, error)
	GetProfileByID(ctx context.Context, id uuid.UUID) (db.Profile, error)
	ListProfilesByUser(ctx context.Context, userID uuid.UUID) ([]db.Profile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error)
	DeleteProfile(ctx context.Context, id uuid.UUID) error
}

type profileRepository struct {
	queries *db.Queries
	db      *sql.DB
}

func NewProfileRepository(queries *db.Queries, db *sql.DB) ProfileRepository {
	return &profileRepository{queries: queries, db: db}
}

func (r *profileRepository) CreateProfile(ctx context.Context, name string, dateOfBirth sql.NullTime, timezone string) (db.Profile, error) {
	return r.queries.CreateProfile(ctx, db.CreateProfileParams{
		Name:        name,
		DateOfBirth: dateOfBirth,
		Timezone:    timezone,
	})
}

func (r *profileRepository) CreateProfilePermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permissions json.RawMessage) error {
	_, err := r.queries.CreateProfilePermission(ctx, db.CreateProfilePermissionParams{
		ProfileID:        profileID,
		SharedWithUserID: userID,
		GrantedByUserID:  userID,
		Permissions:      permissions,
		Status:           "accepted",
		ExpiresAt:        sql.NullTime{},
	})
	return err
}

func (r *profileRepository) CreateProfileWithPermission(ctx context.Context, name string, dateOfBirth sql.NullTime, timezone string, userID uuid.UUID, permissions json.RawMessage) (db.Profile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return db.Profile{}, err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)

	profile, err := qtx.CreateProfile(ctx, db.CreateProfileParams{
		Name:        name,
		DateOfBirth: dateOfBirth,
		Timezone:    timezone,
	})
	if err != nil {
		return db.Profile{}, err
	}

	_, err = qtx.CreateProfilePermission(ctx, db.CreateProfilePermissionParams{
		ProfileID:        profile.ID,
		SharedWithUserID: userID,
		GrantedByUserID:  userID,
		Permissions:      permissions,
		Status:           "accepted",
		ExpiresAt:        sql.NullTime{},
	})
	if err != nil {
		return db.Profile{}, err
	}

	if err := tx.Commit(); err != nil {
		return db.Profile{}, err
	}

	return profile, nil
}

func (r *profileRepository) GetProfileByID(ctx context.Context, id uuid.UUID) (db.Profile, error) {
	return r.queries.GetProfileByID(ctx, id)
}

func (r *profileRepository) ListProfilesByUser(ctx context.Context, userID uuid.UUID) ([]db.Profile, error) {
	return r.queries.ListProfilesByUser(ctx, userID)
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
