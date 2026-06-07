package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

type DoseScheduleRepository interface {
	CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, name string, t time.Time) (db.DoseSchedule, error)
	GetDoseScheduleByID(ctx context.Context, id uuid.UUID) (db.DoseSchedule, error)
	ListDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) ([]db.DoseSchedule, error)
	UpdateDoseSchedule(ctx context.Context, id uuid.UUID, name string, t time.Time) (db.DoseSchedule, error)
	DeleteDoseSchedule(ctx context.Context, id uuid.UUID) error
	DeleteDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) error
}

type doseScheduleRepository struct {
	queries *db.Queries
}

func NewDoseScheduleRepository(queries *db.Queries) DoseScheduleRepository {
	return &doseScheduleRepository{queries: queries}
}

func (r *doseScheduleRepository) CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, name string, t time.Time) (db.DoseSchedule, error) {
	return r.queries.CreateDoseSchedule(ctx, db.CreateDoseScheduleParams{
		ProfileID: profileID,
		Name:      name,
		Time:      t,
	})
}

func (r *doseScheduleRepository) GetDoseScheduleByID(ctx context.Context, id uuid.UUID) (db.DoseSchedule, error) {
	return r.queries.GetDoseScheduleByID(ctx, id)
}

func (r *doseScheduleRepository) ListDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) ([]db.DoseSchedule, error) {
	return r.queries.ListDoseSchedulesByProfile(ctx, profileID)
}

func (r *doseScheduleRepository) UpdateDoseSchedule(ctx context.Context, id uuid.UUID, name string, t time.Time) (db.DoseSchedule, error) {
	return r.queries.UpdateDoseSchedule(ctx, db.UpdateDoseScheduleParams{
		ID:   id,
		Name: name,
		Time: t,
	})
}

func (r *doseScheduleRepository) DeleteDoseSchedule(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteDoseSchedule(ctx, id)
}

func (r *doseScheduleRepository) DeleteDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) error {
	return r.queries.DeleteDoseSchedulesByProfile(ctx, profileID)
}
