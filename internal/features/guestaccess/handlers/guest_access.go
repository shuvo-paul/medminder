package handlers

import (
	"context"
	"database/sql"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/dto"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/service"
	profileService "github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/shuvo-paul/medminder/internal/middleware"
)

func CreateGuestAccessHandler(svc service.GuestAccessService) func(context.Context, *dto.CreateGuestAccessInput) (*dto.CreateGuestAccessOutput, error) {
	return func(ctx context.Context, input *dto.CreateGuestAccessInput) (*dto.CreateGuestAccessOutput, error) {
		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		var label string
		if input.Body.Label != nil {
			label = *input.Body.Label
		}

		expiresInDays := service.DefaultExpiryDays
		if input.Body.ExpiresInDays != nil && *input.Body.ExpiresInDays > 0 {
			expiresInDays = *input.Body.ExpiresInDays
		}

		permissions := input.Body.Permissions

		result, err := svc.CreateToken(ctx, profileID, label, permissions, expiresInDays)
		if err != nil {
			if errors.Is(err, service.ErrInvalidPermissions) {
				return nil, huma.Error400BadRequest("Invalid permissions", err)
			}
			return nil, huma.Error500InternalServerError("Failed to create guest access token", err)
		}

		resp := &dto.CreateGuestAccessOutput{}
		resp.Body.ID = result.ID
		resp.Body.Token = result.RawToken
		resp.Body.ExpiresAt = result.ExpiresAt
		resp.Body.Label = result.Label
		return resp, nil
	}
}

func ListGuestAccessTokensHandler(svc service.GuestAccessService) func(context.Context, *dto.ListGuestAccessTokensInput) (*dto.ListGuestAccessTokensOutput, error) {
	return func(ctx context.Context, input *dto.ListGuestAccessTokensInput) (*dto.ListGuestAccessTokensOutput, error) {
		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		results, err := svc.ListTokens(ctx, profileID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to list guest access tokens", err)
		}

		resp := &dto.ListGuestAccessTokensOutput{}
		resp.Body.Tokens = make([]dto.GuestAccessTokenDTO, len(results))
		for i, r := range results {
			resp.Body.Tokens[i] = dto.GuestAccessTokenDTO{
				ID:         r.ID,
				Label:      r.Label,
				ExpiresAt:  r.ExpiresAt,
				CreatedAt:  r.CreatedAt,
				LastUsedAt: r.LastUsedAt,
			}
		}
		return resp, nil
	}
}

func RevokeGuestAccessHandler(svc service.GuestAccessService) func(context.Context, *dto.RevokeGuestAccessInput) (*dto.RevokeGuestAccessOutput, error) {
	return func(ctx context.Context, input *dto.RevokeGuestAccessInput) (*dto.RevokeGuestAccessOutput, error) {
		userID := middleware.UserIDFromContext(ctx)
		if userID == uuid.Nil {
			return nil, huma.Error401Unauthorized("Invalid or missing user ID", nil)
		}

		tokenID, err := uuid.Parse(input.TokenID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid token ID", err)
		}

		err = svc.RevokeToken(ctx, tokenID, userID)
		if err != nil {
			if errors.Is(err, service.ErrGuestTokenNotFound) {
				return nil, huma.Error404NotFound("Guest access token not found", err)
			}
			if errors.Is(err, service.ErrGuestTokenInsufficientPerms) {
				return nil, huma.Error403Forbidden("Insufficient permissions to revoke this token", err)
			}
			return nil, huma.Error500InternalServerError("Failed to revoke guest access token", err)
		}

		resp := &dto.RevokeGuestAccessOutput{}
		resp.Body.Message = "Guest access token revoked successfully"
		return resp, nil
	}
}

type GuestAuthService interface {
	Authenticate(ctx context.Context, rawToken string) (*service.AuthenticatedToken, error)
}

func GuestListMedicationsHandler(authSvc GuestAuthService, scheduleRepo profileService.DoseScheduleQuerier) func(context.Context, *dto.GuestMedicationInput) (*dto.GuestMedicationsOutput, error) {
	return func(ctx context.Context, input *dto.GuestMedicationInput) (*dto.GuestMedicationsOutput, error) {
		authToken, err := authSvc.Authenticate(ctx, input.Token)
		if err != nil {
			if errors.Is(err, service.ErrGuestTokenNotFound) || errors.Is(err, service.ErrGuestTokenExpired) {
				return nil, huma.Error404NotFound("Not found", nil)
			}
			return nil, huma.Error500InternalServerError("Authentication failed", err)
		}

		if !hasPermission(authToken.Permissions, "medication:read") {
			return nil, huma.Error404NotFound("Not found", nil)
		}

		schedules, err := scheduleRepo.ListDoseSchedulesByProfile(ctx, authToken.ProfileID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to list medications", err)
		}

		resp := &dto.GuestMedicationsOutput{}
		resp.Body.Medications = make([]dto.GuestMedicationDTO, len(schedules))
		for i, s := range schedules {
			resp.Body.Medications[i] = dto.GuestMedicationDTO{
				ID:        s.ID,
				Name:      s.Name,
				Time:      s.Time.Format("15:04"),
				CreatedAt: s.CreatedAt,
				UpdatedAt: s.UpdatedAt,
			}
		}
		return resp, nil
	}
}

func GuestGetReminderHandler(authSvc GuestAuthService, scheduleRepo profileService.DoseScheduleQuerier) func(context.Context, *dto.GuestReminderInput) (*dto.GuestReminderOutput, error) {
	return func(ctx context.Context, input *dto.GuestReminderInput) (*dto.GuestReminderOutput, error) {
		authToken, err := authSvc.Authenticate(ctx, input.Token)
		if err != nil {
			if errors.Is(err, service.ErrGuestTokenNotFound) || errors.Is(err, service.ErrGuestTokenExpired) {
				return nil, huma.Error404NotFound("Not found", nil)
			}
			return nil, huma.Error500InternalServerError("Authentication failed", err)
		}

		if !hasPermission(authToken.Permissions, "reminder:read") {
			return nil, huma.Error404NotFound("Not found", nil)
		}

		reminderID, err := uuid.Parse(input.ReminderID)
		if err != nil {
			return nil, huma.Error404NotFound("Not found", nil)
		}

		schedule, err := scheduleRepo.GetDoseScheduleByID(ctx, reminderID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, huma.Error404NotFound("Not found", nil)
			}
			return nil, huma.Error500InternalServerError("Failed to get reminder", err)
		}

		if schedule.ProfileID != authToken.ProfileID {
			return nil, huma.Error404NotFound("Not found", nil)
		}

		resp := &dto.GuestReminderOutput{}
		resp.Body.Reminder = dto.GuestReminderDTO{
			ID:        schedule.ID,
			Name:      schedule.Name,
			Time:      schedule.Time.Format("15:04"),
			CreatedAt: schedule.CreatedAt,
			UpdatedAt: schedule.UpdatedAt,
		}
		return resp, nil
	}
}

func hasPermission(permissions []string, required string) bool {
	for _, p := range permissions {
		if p == required {
			return true
		}
	}
	return false
}
