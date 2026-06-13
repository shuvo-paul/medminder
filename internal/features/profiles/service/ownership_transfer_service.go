package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/profiles/repository"
)

type OwnershipTransferDTO struct {
	ID          uuid.UUID
	ProfileID   uuid.UUID
	ProfileName string
	FromUserID  uuid.UUID
	FromName    string
	ToUserID    uuid.UUID
	ToName      string
	Status      string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

type OwnershipTransferResult struct {
	Transfer OwnershipTransferDTO
}

type OwnershipTransferService interface {
	InitiateTransfer(ctx context.Context, profileID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID) (*OwnershipTransferResult, error)
	ListPendingTransfers(ctx context.Context, userID uuid.UUID) ([]OwnershipTransferResult, error)
	AcceptTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
	DeclineTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
	CancelTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
}

type ownershipTransferService struct {
	profileRepo  repository.ProfileRepository
	transferRepo repository.OwnershipTransferRepository
}

func NewOwnershipTransferService(profileRepo repository.ProfileRepository, transferRepo repository.OwnershipTransferRepository) OwnershipTransferService {
	return &ownershipTransferService{
		profileRepo:  profileRepo,
		transferRepo: transferRepo,
	}
}

func (s *ownershipTransferService) InitiateTransfer(ctx context.Context, profileID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID) (*OwnershipTransferResult, error) {
	profile, err := s.profileRepo.GetProfileByID(ctx, profileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	if toUserID == fromUserID {
		return nil, ErrCannotTransferToSelf
	}

	exists, err := s.profileRepo.UserExists(ctx, toUserID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	pp, err := s.profileRepo.GetProfilePermission(ctx, profileID, toUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNewOwnerNoAdminPermission
		}
		return nil, err
	}

	if pp.Status != "accepted" {
		return nil, ErrNewOwnerNoAdminPermission
	}

	var perms []string
	if err := json.Unmarshal(pp.Permissions, &perms); err != nil {
		return nil, err
	}

	if !slices.Contains(perms, "profile:admin") {
		return nil, ErrNewOwnerNoAdminPermission
	}

	_, err = s.transferRepo.GetPendingTransferByProfile(ctx, profileID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		return nil, ErrPendingTransferExists
	}

	expiresAt := time.Now().AddDate(0, 0, 7)
	transfer, err := s.transferRepo.CreateTransfer(ctx, profileID, fromUserID, toUserID, expiresAt)
	if err != nil {
		return nil, err
	}

	fromUser, err := s.profileRepo.GetUserByID(ctx, fromUserID)
	if err != nil {
		return nil, err
	}

	toUser, err := s.profileRepo.GetUserByID(ctx, toUserID)
	if err != nil {
		return nil, err
	}

	return &OwnershipTransferResult{
		Transfer: OwnershipTransferDTO{
			ID:          transfer.ID,
			ProfileID:   transfer.ProfileID,
			ProfileName: profile.Name,
			FromUserID:  transfer.FromUserID,
			FromName:    fromUser.DisplayName,
			ToUserID:    transfer.ToUserID,
			ToName:      toUser.DisplayName,
			Status:      transfer.Status,
			ExpiresAt:   transfer.ExpiresAt,
			CreatedAt:   transfer.CreatedAt,
		},
	}, nil
}

func (s *ownershipTransferService) ListPendingTransfers(ctx context.Context, userID uuid.UUID) ([]OwnershipTransferResult, error) {
	transfers, err := s.transferRepo.ListPendingTransfersWithDetailsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results []OwnershipTransferResult
	for _, t := range transfers {
		results = append(results, OwnershipTransferResult{
			Transfer: OwnershipTransferDTO{
				ID:          t.ID,
				ProfileID:   t.ProfileID,
				ProfileName: t.ProfileName,
				FromUserID:  t.FromUserID,
				FromName:    t.FromName,
				ToUserID:    t.ToUserID,
				ToName:      t.ToName,
				Status:      t.Status,
				ExpiresAt:   t.ExpiresAt,
				CreatedAt:   t.CreatedAt,
			},
		})
	}

	return results, nil
}

func (s *ownershipTransferService) AcceptTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	transfer, err := s.transferRepo.GetTransferByID(ctx, transferID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTransferNotFound
		}
		return err
	}

	if transfer.ToUserID != userID {
		return ErrTransferNotFound
	}

	if transfer.Status != "pending" {
		return ErrTransferNotPending
	}

	if transfer.ExpiresAt.Before(time.Now()) {
		return ErrTransferExpired
	}

	newOwnerPerm, err := s.profileRepo.GetProfilePermission(ctx, transfer.ProfileID, transfer.ToUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNewOwnerNoAdminPermission
		}
		return err
	}

	var newPerms []string
	if err := json.Unmarshal(newOwnerPerm.Permissions, &newPerms); err != nil {
		return err
	}

	if !slices.Contains(newPerms, "profile:admin") {
		return ErrNewOwnerNoAdminPermission
	}

	newPerms = append(newPerms, "profile:owner")
	newPermsJSON, err := json.Marshal(newPerms)
	if err != nil {
		return err
	}

	oldOwnerPerm, err := s.profileRepo.GetProfilePermission(ctx, transfer.ProfileID, transfer.FromUserID)
	if err != nil {
		return err
	}

	var oldPerms []string
	if err := json.Unmarshal(oldOwnerPerm.Permissions, &oldPerms); err != nil {
		return err
	}

	oldPerms = removeString(oldPerms, "profile:owner")
	oldPermsJSON, err := json.Marshal(oldPerms)
	if err != nil {
		return err
	}

	if _, err := s.profileRepo.UpdateProfilePermissionByProfileAndUser(ctx, transfer.ProfileID, transfer.FromUserID, oldPermsJSON); err != nil {
		return err
	}

	if _, err := s.profileRepo.UpdateProfilePermissionByProfileAndUser(ctx, transfer.ProfileID, transfer.ToUserID, newPermsJSON); err != nil {
		return err
	}

	if _, err := s.transferRepo.UpdateTransferStatus(ctx, transferID, "accepted"); err != nil {
		return err
	}

	return nil
}

func (s *ownershipTransferService) DeclineTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	transfer, err := s.transferRepo.GetTransferByID(ctx, transferID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTransferNotFound
		}
		return err
	}

	if transfer.ToUserID != userID {
		return ErrTransferNotFound
	}

	if transfer.Status != "pending" {
		return ErrTransferNotPending
	}

	if transfer.ExpiresAt.Before(time.Now()) {
		return ErrTransferExpired
	}

	_, err = s.transferRepo.UpdateTransferStatus(ctx, transferID, "declined")
	return err
}

func (s *ownershipTransferService) CancelTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	transfer, err := s.transferRepo.GetTransferByID(ctx, transferID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrTransferNotFound
		}
		return err
	}

	if transfer.FromUserID != userID {
		return ErrTransferNotInitiator
	}

	if transfer.Status != "pending" {
		return ErrTransferNotPending
	}

	if transfer.ExpiresAt.Before(time.Now()) {
		return ErrTransferExpired
	}

	_, err = s.transferRepo.UpdateTransferStatus(ctx, transferID, "cancelled")
	return err
}

func removeString(slice []string, s string) []string {
	var result []string
	for _, v := range slice {
		if v != s {
			result = append(result, v)
		}
	}
	return result
}
