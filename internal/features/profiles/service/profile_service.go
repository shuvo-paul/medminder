package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/repository"
)

type DoseScheduleInput struct {
	Name string
	Time string
}

type ProfileService interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, name string, dateOfBirth *time.Time, timezone string, schedules []DoseScheduleInput) (*dto.ProfileDTO, error)
	GetProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) (*dto.ProfileDTO, error)
	ListProfiles(ctx context.Context, userID uuid.UUID) ([]dto.ProfileDTO, error)
	UpdateProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name *string, dateOfBirth *time.Time, timezone *string) (*dto.ProfileDTO, error)
	DeleteProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) error
	HandleAccountDeletion(ctx context.Context, userID uuid.UUID) error
}

type profileService struct {
	profileRepo  repository.ProfileRepository
	permRepo     repository.PermissionRepository
	scheduleRepo DoseScheduleQuerier
	permChecker  PermissionChecker
}

func NewProfileService(profileRepo repository.ProfileRepository, permRepo repository.PermissionRepository, scheduleRepo DoseScheduleQuerier, permChecker PermissionChecker) ProfileService {
	return &profileService{
		profileRepo:  profileRepo,
		permRepo:     permRepo,
		scheduleRepo: scheduleRepo,
		permChecker:  permChecker,
	}
}

func (s *profileService) CreateProfile(ctx context.Context, userID uuid.UUID, name string, dateOfBirth *time.Time, timezone string, schedules []DoseScheduleInput) (*dto.ProfileDTO, error) {
	if _, err := time.LoadLocation(timezone); err != nil {
		return nil, ErrInvalidTimezone
	}

	var dob sql.NullTime
	if dateOfBirth != nil {
		dob = sql.NullTime{Time: *dateOfBirth, Valid: true}
	}

	ownerPerms, err := json.Marshal([]string{"profile:owner", "profile:admin"})
	if err != nil {
		return nil, err
	}

	profile, err := s.profileRepo.CreateProfileWithPermission(ctx, name, dob, timezone, userID, ownerPerms)
	if err != nil {
		return nil, err
	}

	for _, schedule := range schedules {
		t, err := time.Parse("15:04", schedule.Time)
		if err != nil {
			continue
		}
		_, err = s.scheduleRepo.CreateDoseSchedule(ctx, profile.ID, schedule.Name, t)
		if err != nil {
			return nil, err
		}
	}

	profileDTO := toProfileDTO(profile, []dto.DoseScheduleDTO{}, true)
	return &profileDTO, nil
}

func (s *profileService) GetProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) (*dto.ProfileDTO, error) {
	profile, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	isOwner, err := s.permChecker.HasPermission(ctx, profileID, userID, "profile:owner")
	if err != nil {
		return nil, err
	}

	schedules, err := s.scheduleRepo.ListDoseSchedulesByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}

	profileDTO := toProfileDTO(profile, toDTOs(schedules), isOwner)
	return &profileDTO, nil
}

func (s *profileService) ListProfiles(ctx context.Context, userID uuid.UUID) ([]dto.ProfileDTO, error) {
	profiles, err := s.profileRepo.ListProfilesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results []dto.ProfileDTO
	for _, profile := range profiles {
		isOwner, err := s.permChecker.HasPermission(ctx, profile.ID, userID, "profile:owner")
		if err != nil {
			return nil, err
		}
		schedules, err := s.scheduleRepo.ListDoseSchedulesByProfile(ctx, profile.ID)
		if err != nil {
			return nil, err
		}
		results = append(results, toProfileDTO(profile, toDTOs(schedules), isOwner))
	}

	return results, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name *string, dateOfBirth *time.Time, timezone *string) (*dto.ProfileDTO, error) {
	profile, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	isOwner, err := s.permChecker.HasPermission(ctx, profileID, userID, "profile:owner")
	if err != nil {
		return nil, err
	}

	updatedName := profile.Name
	if name != nil {
		updatedName = *name
	}

	updatedTimezone := profile.Timezone
	if timezone != nil {
		if _, err := time.LoadLocation(*timezone); err != nil {
			return nil, ErrInvalidTimezone
		}
		updatedTimezone = *timezone
	}

	var dob sql.NullTime
	if dateOfBirth != nil {
		dob = sql.NullTime{Time: *dateOfBirth, Valid: true}
	} else {
		dob = profile.DateOfBirth
	}

	updated, err := s.profileRepo.UpdateProfile(ctx, profileID, updatedName, dob, updatedTimezone)
	if err != nil {
		return nil, err
	}

	schedules, err := s.scheduleRepo.ListDoseSchedulesByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}

	profileDTO := toProfileDTO(updated, toDTOs(schedules), isOwner)
	return &profileDTO, nil
}

func (s *profileService) DeleteProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) error {
	_, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrProfileNotFound
		}
		return err
	}

	if err := s.scheduleRepo.DeleteDoseSchedulesByProfile(ctx, profileID); err != nil {
		return err
	}

	return s.profileRepo.DeleteProfile(ctx, profileID)
}

func (s *profileService) HandleAccountDeletion(ctx context.Context, userID uuid.UUID) error {
	permissions, err := s.permRepo.ListProfilePermissionsByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, perm := range permissions {
		profileID := perm.ProfileID

		if HasPermissionInJSONB(perm.Permissions, "profile:owner") {
			profilePerms, err := s.permRepo.ListProfilePermissionsByProfile(ctx, profileID)
			if err != nil {
				return err
			}

			hasOtherAdmin := false
			for _, pp := range profilePerms {
				if pp.SharedWithUserID != userID &&
					pp.Status == "accepted" &&
					HasPermissionInJSONB(pp.Permissions, "profile:admin") {
					hasOtherAdmin = true
					break
				}
			}

			if hasOtherAdmin {
				if err := s.permRepo.DeleteProfilePermission(ctx, perm.ID); err != nil {
					return err
				}
			} else {
				if err := s.profileRepo.DeleteProfile(ctx, profileID); err != nil {
					return err
				}
			}
		} else {
			if err := s.permRepo.DeleteProfilePermission(ctx, perm.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func toProfileDTO(profile db.Profile, schedules []dto.DoseScheduleDTO, isOwner bool) dto.ProfileDTO {
	var dob *string
	if profile.DateOfBirth.Valid {
		s := profile.DateOfBirth.Time.Format("2006-01-02")
		dob = &s
	}

	return dto.ProfileDTO{
		ID:          profile.ID,
		Name:        profile.Name,
		DateOfBirth: dob,
		Timezone:    profile.Timezone,
		IsOwner:     isOwner,
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
		Schedules:   schedules,
	}
}

func toDTOs(schedules []db.DoseSchedule) []dto.DoseScheduleDTO {
	result := make([]dto.DoseScheduleDTO, len(schedules))
	for i, s := range schedules {
		result[i] = dto.DoseScheduleDTO{
			ID:        s.ID,
			ProfileID: s.ProfileID,
			Name:      s.Name,
			Time:      s.Time.Format("15:04"),
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		}
	}
	return result
}
