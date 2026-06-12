package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/profiles/repository"
)

var validSharePermissions = []string{"profile:read", "profile:write", "profile:admin", "profile:share"}

type InvitationDTO struct {
	ID               uuid.UUID
	ProfileID        uuid.UUID
	ProfileName      string
	SharedWithUserID uuid.UUID
	GrantedByUserID  uuid.UUID
	Permissions      []string
	Status           string
	ExpiresAt        *time.Time
	CreatedAt        time.Time
}

type ShareInput struct {
	SharedWithUserID uuid.UUID
	Permissions      []string
	ExpiresInDays    int
}

type InvitationResult struct {
	Invitation InvitationDTO
}

type AcceptedProfileResult struct {
	Profile     ProfileDTO
	Permissions []string
}

type InvitationService interface {
	ShareProfile(ctx context.Context, profileID uuid.UUID, grantedByUserID uuid.UUID, input ShareInput) (*InvitationResult, error)
	ListInvitations(ctx context.Context, userID uuid.UUID) ([]InvitationResult, error)
	AcceptInvitation(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) (*AcceptedProfileResult, error)
	DeclineInvitation(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error
}

type invitationService struct {
	profileRepo  repository.ProfileRepository
	scheduleRepo repository.DoseScheduleRepository
}

func NewInvitationService(profileRepo repository.ProfileRepository, scheduleRepo repository.DoseScheduleRepository) InvitationService {
	return &invitationService{
		profileRepo:  profileRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (s *invitationService) ShareProfile(ctx context.Context, profileID uuid.UUID, grantedByUserID uuid.UUID, input ShareInput) (*InvitationResult, error) {
	if _, err := s.profileRepo.GetProfileByID(ctx, profileID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	if input.SharedWithUserID == grantedByUserID {
		return nil, ErrCannotShareWithSelf
	}

	for _, p := range input.Permissions {
		if p == "profile:owner" || !slices.Contains(validSharePermissions, p) {
			return nil, ErrInvalidPermissions
		}
	}

	if len(input.Permissions) == 0 {
		return nil, ErrInvalidPermissions
	}

	exists, err := s.profileRepo.UserExists(ctx, input.SharedWithUserID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	existing, err := s.profileRepo.GetProfilePermission(ctx, profileID, input.SharedWithUserID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err != sql.ErrNoRows {
		if existing.Status == "accepted" || existing.Status == "pending" {
			return nil, ErrUserAlreadySharing
		}
	}

	permsJSON, err := json.Marshal(input.Permissions)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().AddDate(0, 0, input.ExpiresInDays)

	pp, err := s.profileRepo.CreateInvitation(ctx, profileID, input.SharedWithUserID, grantedByUserID, permsJSON, expiresAt)
	if err != nil {
		return nil, err
	}

	result, err := toInvitationResult(pp, "")
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *invitationService) ListInvitations(ctx context.Context, userID uuid.UUID) ([]InvitationResult, error) {
	permissions, err := s.profileRepo.ListProfilePermissionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results []InvitationResult
	now := time.Now()

	for _, pp := range permissions {
		if pp.Status != "pending" {
			continue
		}

		if pp.ExpiresAt.Valid && pp.ExpiresAt.Time.Before(now) {
			continue
		}

		profile, err := s.profileRepo.GetProfileByID(ctx, pp.ProfileID)
		if err != nil {
			return nil, err
		}

		result, err := toInvitationResult(pp, profile.Name)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (s *invitationService) AcceptInvitation(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) (*AcceptedProfileResult, error) {
	pp, err := s.profileRepo.GetProfilePermissionByID(ctx, invitationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	if pp.SharedWithUserID != userID {
		return nil, ErrInvitationNotFound
	}

	if pp.Status != "pending" {
		return nil, ErrInvitationAlreadyProcessed
	}

	if pp.ExpiresAt.Valid && pp.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrInvitationExpired
	}

	profile, err := s.profileRepo.GetProfileByID(ctx, pp.ProfileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	accepted, err := s.profileRepo.AcceptProfilePermission(ctx, invitationID)
	if err != nil {
		return nil, err
	}

	var perms []string
	if err := json.Unmarshal(accepted.Permissions, &perms); err != nil {
		return nil, err
	}

	schedules, err := s.scheduleRepo.ListDoseSchedulesByProfile(ctx, pp.ProfileID)
	if err != nil {
		return nil, err
	}

	isOwner := false
	for _, p := range perms {
		if p == "profile:owner" {
			isOwner = true
			break
		}
	}

	profileResult := toProfileResult(profile, toDoseScheduleDTOs(schedules), isOwner)

	return &AcceptedProfileResult{
		Profile:     profileResult.Profile,
		Permissions: perms,
	}, nil
}

func (s *invitationService) DeclineInvitation(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	pp, err := s.profileRepo.GetProfilePermissionByID(ctx, invitationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvitationNotFound
		}
		return err
	}

	if pp.SharedWithUserID != userID {
		return ErrInvitationNotFound
	}

	if pp.Status != "pending" {
		return ErrInvitationAlreadyProcessed
	}

	_, err = s.profileRepo.UpdateProfilePermissionStatus(ctx, invitationID, "declined")
	return err
}

func toInvitationResult(pp db.ProfilePermission, profileName string) (InvitationResult, error) {
	var perms []string
	if err := json.Unmarshal(pp.Permissions, &perms); err != nil {
		return InvitationResult{}, err
	}

	var expiresAt *time.Time
	if pp.ExpiresAt.Valid {
		expiresAt = &pp.ExpiresAt.Time
	}

	return InvitationResult{
		Invitation: InvitationDTO{
			ID:               pp.ID,
			ProfileID:        pp.ProfileID,
			ProfileName:      profileName,
			SharedWithUserID: pp.SharedWithUserID,
			GrantedByUserID:  pp.GrantedByUserID,
			Permissions:      perms,
			Status:           pp.Status,
			ExpiresAt:        expiresAt,
			CreatedAt:        pp.CreatedAt,
		},
	}, nil
}
