package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
)

type OwnershipTransferServiceInterface interface {
	InitiateTransfer(ctx context.Context, profileID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID) (*service.OwnershipTransferResult, error)
	ListPendingTransfers(ctx context.Context, userID uuid.UUID) ([]service.OwnershipTransferResult, error)
	AcceptTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
	DeclineTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
	CancelTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error
}

func InitiateTransferHandler(svc OwnershipTransferServiceInterface, tokenSvc TokenServiceInterface) func(context.Context, *dto.InitiateTransferInput) (*dto.InitiateTransferOutput, error) {
	return func(ctx context.Context, input *dto.InitiateTransferInput) (*dto.InitiateTransferOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		result, err := svc.InitiateTransfer(ctx, profileID, userID, input.Body.ToUserID)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrCannotTransferToSelf) {
				return nil, huma.Error400BadRequest("Cannot transfer ownership to yourself", err)
			}
			if errors.Is(err, service.ErrUserNotFound) {
				return nil, huma.Error404NotFound("User not found", err)
			}
			if errors.Is(err, service.ErrNewOwnerNoAdminPermission) {
				return nil, huma.Error400BadRequest("New owner must have profile:admin permission", err)
			}
			if errors.Is(err, service.ErrPendingTransferExists) {
				return nil, huma.Error409Conflict("A pending transfer already exists for this profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to initiate transfer", err)
		}

		resp := &dto.InitiateTransferOutput{}
		resp.Body.Transfer = toOwnershipTransferDTO(result.Transfer)
		return resp, nil
	}
}

func ListTransfersHandler(svc OwnershipTransferServiceInterface, tokenSvc TokenServiceInterface) func(context.Context, *dto.ListTransfersInput) (*dto.ListTransfersOutput, error) {
	return func(ctx context.Context, input *dto.ListTransfersInput) (*dto.ListTransfersOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		results, err := svc.ListPendingTransfers(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to list transfers", err)
		}

		resp := &dto.ListTransfersOutput{}
		resp.Body.Transfers = make([]dto.OwnershipTransferDTO, len(results))
		for i, r := range results {
			resp.Body.Transfers[i] = toOwnershipTransferDTO(r.Transfer)
		}
		return resp, nil
	}
}

func AcceptTransferHandler(svc OwnershipTransferServiceInterface, tokenSvc TokenServiceInterface) func(context.Context, *dto.TransferActionInput) (*dto.TransferActionOutput, error) {
	return func(ctx context.Context, input *dto.TransferActionInput) (*dto.TransferActionOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		transferID, err := uuid.Parse(input.TransferID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid transfer ID", err)
		}

		err = svc.AcceptTransfer(ctx, transferID, userID)
		if err != nil {
			if errors.Is(err, service.ErrTransferNotFound) {
				return nil, huma.Error404NotFound("Transfer not found", err)
			}
			if errors.Is(err, service.ErrTransferNotPending) {
				return nil, huma.Error409Conflict("Transfer is not pending", err)
			}
			if errors.Is(err, service.ErrTransferExpired) {
				return nil, huma.Error400BadRequest("Transfer has expired", err)
			}
			if errors.Is(err, service.ErrNewOwnerNoAdminPermission) {
				return nil, huma.Error400BadRequest("You no longer have profile:admin permission on this profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to accept transfer", err)
		}

		resp := &dto.TransferActionOutput{}
		resp.Body.Message = "Transfer accepted successfully"
		return resp, nil
	}
}

func DeclineTransferHandler(svc OwnershipTransferServiceInterface, tokenSvc TokenServiceInterface) func(context.Context, *dto.TransferActionInput) (*dto.TransferActionOutput, error) {
	return func(ctx context.Context, input *dto.TransferActionInput) (*dto.TransferActionOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		transferID, err := uuid.Parse(input.TransferID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid transfer ID", err)
		}

		err = svc.DeclineTransfer(ctx, transferID, userID)
		if err != nil {
			if errors.Is(err, service.ErrTransferNotFound) {
				return nil, huma.Error404NotFound("Transfer not found", err)
			}
			return nil, huma.Error500InternalServerError("Failed to decline transfer", err)
		}

		resp := &dto.TransferActionOutput{}
		resp.Body.Message = "Transfer declined successfully"
		return resp, nil
	}
}

func CancelTransferHandler(svc OwnershipTransferServiceInterface, tokenSvc TokenServiceInterface) func(context.Context, *dto.TransferActionInput) (*dto.TransferActionOutput, error) {
	return func(ctx context.Context, input *dto.TransferActionInput) (*dto.TransferActionOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		transferID, err := uuid.Parse(input.TransferID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid transfer ID", err)
		}

		err = svc.CancelTransfer(ctx, transferID, userID)
		if err != nil {
			if errors.Is(err, service.ErrTransferNotFound) {
				return nil, huma.Error404NotFound("Transfer not found", err)
			}
			if errors.Is(err, service.ErrTransferNotInitiator) {
				return nil, huma.Error403Forbidden("Only the transfer initiator can cancel", err)
			}
			return nil, huma.Error500InternalServerError("Failed to cancel transfer", err)
		}

		resp := &dto.TransferActionOutput{}
		resp.Body.Message = "Transfer cancelled successfully"
		return resp, nil
	}
}

func toOwnershipTransferDTO(t service.OwnershipTransferDTO) dto.OwnershipTransferDTO {
	return dto.OwnershipTransferDTO{
		ID:          t.ID,
		ProfileID:   t.ProfileID,
		ProfileName: t.ProfileName,
		FromUserID:  t.FromUserID,
		FromName:    t.FromName,
		ToUserID:    t.ToUserID,
		Status:      t.Status,
		ExpiresAt:   t.ExpiresAt,
		CreatedAt:   t.CreatedAt,
	}
}
