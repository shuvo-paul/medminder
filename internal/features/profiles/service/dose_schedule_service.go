package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/profiles/repository"
)

type DoseScheduleService interface {
	CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name string, timeStr string) (*DoseScheduleResult, error)
	GetDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) (*DoseScheduleResult, error)
	ListDoseSchedules(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) ([]DoseScheduleResult, error)
	UpdateDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID, name *string, timeStr *string) (*DoseScheduleResult, error)
	DeleteDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) error
}

type doseScheduleService struct {
	profileRepo  repository.ProfileRepository
	scheduleRepo repository.DoseScheduleRepository
}

func NewDoseScheduleService(profileRepo repository.ProfileRepository, scheduleRepo repository.DoseScheduleRepository) DoseScheduleService {
	return &doseScheduleService{
		profileRepo:  profileRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (s *doseScheduleService) CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name string, timeStr string) (*DoseScheduleResult, error) {
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

	result := toDoseScheduleResult(schedule)
	return &result, nil
}

func (s *doseScheduleService) GetDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) (*DoseScheduleResult, error) {
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

	result := toDoseScheduleResult(schedule)
	return &result, nil
}

func (s *doseScheduleService) ListDoseSchedules(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) ([]DoseScheduleResult, error) {
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

	return toDoseScheduleResults(schedules), nil
}

func (s *doseScheduleService) UpdateDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID, name *string, timeStr *string) (*DoseScheduleResult, error) {
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

	result := toDoseScheduleResult(updated)
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

type DoseScheduleResult struct {
	Schedule DoseScheduleDTO
}

func toDoseScheduleResult(schedule db.DoseSchedule) DoseScheduleResult {
	return DoseScheduleResult{
		Schedule: DoseScheduleDTO{
			ID:        schedule.ID,
			ProfileID: schedule.ProfileID,
			Name:      schedule.Name,
			Time:      schedule.Time.Format("15:04"),
			CreatedAt: schedule.CreatedAt,
			UpdatedAt: schedule.UpdatedAt,
		},
	}
}

func toDoseScheduleResults(schedules []db.DoseSchedule) []DoseScheduleResult {
	results := make([]DoseScheduleResult, len(schedules))
	for i, s := range schedules {
		results[i] = toDoseScheduleResult(s)
	}
	return results
}
