package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/common/auth"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
)

func ShareProfileHandler(svc service.InvitationService, tokenSvc auth.TokenValidator) func(context.Context, *dto.ShareProfileInput) (*dto.ShareProfileOutput, error) {
	return func(ctx context.Context, input *dto.ShareProfileInput) (*dto.ShareProfileOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		result, err := svc.ShareProfile(ctx, profileID, userID, service.ShareInput{
			SharedWithUserID: input.Body.SharedWithUserID,
			Permissions:      input.Body.Permissions,
			ExpiresInDays:    input.Body.ExpiresInDays,
		})
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrCannotShareWithSelf) {
				return nil, huma.Error400BadRequest("Cannot share profile with yourself", err)
			}
			if errors.Is(err, service.ErrInvalidPermissions) {
				return nil, huma.Error400BadRequest("Invalid permissions", err)
			}
			if errors.Is(err, service.ErrUserNotFound) {
				return nil, huma.Error404NotFound("User not found", err)
			}
			if errors.Is(err, service.ErrUserAlreadySharing) {
				return nil, huma.Error409Conflict("User already has access to this profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to share profile", err)
		}

		resp := &dto.ShareProfileOutput{}
		resp.Body.Invitation = toInvitationDTO(result.Invitation)
		return resp, nil
	}
}

func ListInvitationsHandler(svc service.InvitationService, tokenSvc auth.TokenValidator) func(context.Context, *dto.ListInvitationsInput) (*dto.ListInvitationsOutput, error) {
	return func(ctx context.Context, input *dto.ListInvitationsInput) (*dto.ListInvitationsOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		results, err := svc.ListInvitations(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to list invitations", err)
		}

		resp := &dto.ListInvitationsOutput{}
		resp.Body.Invitations = make([]dto.InvitationDTO, len(results))
		for i, r := range results {
			resp.Body.Invitations[i] = toInvitationDTO(r.Invitation)
		}
		return resp, nil
	}
}

func AcceptInvitationHandler(svc service.InvitationService, tokenSvc auth.TokenValidator) func(context.Context, *dto.AcceptInvitationInput) (*dto.AcceptInvitationOutput, error) {
	return func(ctx context.Context, input *dto.AcceptInvitationInput) (*dto.AcceptInvitationOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		invitationID, err := uuid.Parse(input.InvitationID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid invitation ID", err)
		}

		result, err := svc.AcceptInvitation(ctx, invitationID, userID)
		if err != nil {
			if errors.Is(err, service.ErrInvitationNotFound) {
				return nil, huma.Error404NotFound("Invitation not found", err)
			}
			if errors.Is(err, service.ErrInvitationExpired) {
				return nil, huma.NewError(410, "Invitation has expired", err)
			}
			if errors.Is(err, service.ErrInvitationAlreadyProcessed) {
				return nil, huma.Error409Conflict("Invitation has already been processed", err)
			}
			return nil, huma.Error500InternalServerError("Failed to accept invitation", err)
		}

		resp := &dto.AcceptInvitationOutput{}
		resp.Body.Profile = result.Profile
		resp.Body.Permissions = result.Permissions
		return resp, nil
	}
}

func DeclineInvitationHandler(svc service.InvitationService, tokenSvc auth.TokenValidator) func(context.Context, *dto.DeclineInvitationInput) (*dto.DeclineInvitationOutput, error) {
	return func(ctx context.Context, input *dto.DeclineInvitationInput) (*dto.DeclineInvitationOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		invitationID, err := uuid.Parse(input.InvitationID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid invitation ID", err)
		}

		err = svc.DeclineInvitation(ctx, invitationID, userID)
		if err != nil {
			if errors.Is(err, service.ErrInvitationNotFound) {
				return nil, huma.Error404NotFound("Invitation not found", err)
			}
			if errors.Is(err, service.ErrInvitationAlreadyProcessed) {
				return nil, huma.Error409Conflict("Invitation has already been processed", err)
			}
			return nil, huma.Error500InternalServerError("Failed to decline invitation", err)
		}

		resp := &dto.DeclineInvitationOutput{}
		resp.Body.Message = "Invitation declined successfully"
		return resp, nil
	}
}

func toInvitationDTO(inv service.InvitationDTO) dto.InvitationDTO {
	return dto.InvitationDTO{
		ID:               inv.ID,
		ProfileID:        inv.ProfileID,
		ProfileName:      inv.ProfileName,
		SharedWithUserID: inv.SharedWithUserID,
		GrantedByUserID:  inv.GrantedByUserID,
		Permissions:      inv.Permissions,
		Status:           inv.Status,
		ExpiresAt:        inv.ExpiresAt,
		CreatedAt:        inv.CreatedAt,
	}
}
