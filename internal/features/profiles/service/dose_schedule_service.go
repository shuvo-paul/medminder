package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/repository"
)

type DoseScheduleQuerier interface {
	CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, name string, t time.Time) (db.DoseSchedule, error)
	GetDoseScheduleByID(ctx context.Context, id uuid.UUID) (db.DoseSchedule, error)
	ListDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) ([]db.DoseSchedule, error)
	UpdateDoseSchedule(ctx context.Context, id uuid.UUID, name string, t time.Time) (db.DoseSchedule, error)
	DeleteDoseSchedule(ctx context.Context, id uuid.UUID) error
	DeleteDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) error
}

type doseScheduleQueries struct {
	q *db.Queries
}

func NewDoseScheduleQuerier(q *db.Queries) DoseScheduleQuerier {
	return &doseScheduleQueries{q: q}
}

func (d *doseScheduleQueries) CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, name string, t time.Time) (db.DoseSchedule, error) {
	return d.q.CreateDoseSchedule(ctx, db.CreateDoseScheduleParams{
		ProfileID: profileID,
		Name:      name,
		Time:      t,
	})
}

func (d *doseScheduleQueries) GetDoseScheduleByID(ctx context.Context, id uuid.UUID) (db.DoseSchedule, error) {
	return d.q.GetDoseScheduleByID(ctx, id)
}

func (d *doseScheduleQueries) ListDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) ([]db.DoseSchedule, error) {
	return d.q.ListDoseSchedulesByProfile(ctx, profileID)
}

func (d *doseScheduleQueries) UpdateDoseSchedule(ctx context.Context, id uuid.UUID, name string, t time.Time) (db.DoseSchedule, error) {
	return d.q.UpdateDoseSchedule(ctx, db.UpdateDoseScheduleParams{
		ID:   id,
		Name: name,
		Time: t,
	})
}

func (d *doseScheduleQueries) DeleteDoseSchedule(ctx context.Context, id uuid.UUID) error {
	return d.q.DeleteDoseSchedule(ctx, id)
}

func (d *doseScheduleQueries) DeleteDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) error {
	return d.q.DeleteDoseSchedulesByProfile(ctx, profileID)
}

type DoseScheduleService interface {
	CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name string, timeStr string) (*dto.DoseScheduleDTO, error)
	GetDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) (*dto.DoseScheduleDTO, error)
	ListDoseSchedules(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) ([]dto.DoseScheduleDTO, error)
	UpdateDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID, name *string, timeStr *string) (*dto.DoseScheduleDTO, error)
	DeleteDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) error
}

type doseScheduleService struct {
	profileRepo  repository.ProfileRepository
	scheduleRepo DoseScheduleQuerier
}

func NewDoseScheduleService(profileRepo repository.ProfileRepository, scheduleRepo DoseScheduleQuerier) DoseScheduleService {
	return &doseScheduleService{
		profileRepo:  profileRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (s *doseScheduleService) CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name string, timeStr string) (*dto.DoseScheduleDTO, error) {
	_, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return nil, ErrInvalidTimezone
	}

	schedule, err := s.scheduleRepo.CreateDoseSchedule(ctx, profileID, name, t)
	if err != nil {
		return nil, err
	}

	result := toDoseScheduleDTO(schedule)
	return &result, nil
}

func (s *doseScheduleService) GetDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) (*dto.DoseScheduleDTO, error) {
	_, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	schedule, err := s.scheduleRepo.GetDoseScheduleByID(ctx, scheduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	if schedule.ProfileID != profileID {
		return nil, ErrScheduleNotFound
	}

	result := toDoseScheduleDTO(schedule)
	return &result, nil
}

func (s *doseScheduleService) ListDoseSchedules(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) ([]dto.DoseScheduleDTO, error) {
	_, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	schedules, err := s.scheduleRepo.ListDoseSchedulesByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}

	return toDoseScheduleDTOs(schedules), nil
}

func (s *doseScheduleService) UpdateDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID, name *string, timeStr *string) (*dto.DoseScheduleDTO, error) {
	_, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	schedule, err := s.scheduleRepo.GetDoseScheduleByID(ctx, scheduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	if schedule.ProfileID != profileID {
		return nil, ErrScheduleNotFound
	}

	updatedName := schedule.Name
	if name != nil {
		updatedName = *name
	}

	updatedTime := schedule.Time
	if timeStr != nil {
		t, err := time.Parse("15:04", *timeStr)
		if err != nil {
			return nil, ErrInvalidTimezone
		}
		updatedTime = t
	}

	updated, err := s.scheduleRepo.UpdateDoseSchedule(ctx, scheduleID, updatedName, updatedTime)
	if err != nil {
		return nil, err
	}

	result := toDoseScheduleDTO(updated)
	return &result, nil
}

func (s *doseScheduleService) DeleteDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) error {
	_, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrProfileNotFound
		}
		return err
	}

	schedule, err := s.scheduleRepo.GetDoseScheduleByID(ctx, scheduleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrScheduleNotFound
		}
		return err
	}

	if schedule.ProfileID != profileID {
		return ErrScheduleNotFound
	}

	return s.scheduleRepo.DeleteDoseSchedule(ctx, scheduleID)
}

func toDoseScheduleDTO(schedule db.DoseSchedule) dto.DoseScheduleDTO {
	return dto.DoseScheduleDTO{
		ID:        schedule.ID,
		ProfileID: schedule.ProfileID,
		Name:      schedule.Name,
		Time:      schedule.Time.Format("15:04"),
		CreatedAt: schedule.CreatedAt,
		UpdatedAt: schedule.UpdatedAt,
	}
}

func toDoseScheduleDTOs(schedules []db.DoseSchedule) []dto.DoseScheduleDTO {
	results := make([]dto.DoseScheduleDTO, len(schedules))
	for i, s := range schedules {
		results[i] = toDoseScheduleDTO(s)
	}
	return results
}
