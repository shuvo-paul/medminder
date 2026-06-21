package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/common/auth"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
)

func CreateProfileHandler(svc service.ProfileService, tokenSvc auth.TokenValidator) func(context.Context, *dto.CreateProfileInput) (*dto.CreateProfileOutput, error) {
	return func(ctx context.Context, input *dto.CreateProfileInput) (*dto.CreateProfileOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		var dateOfBirth *time.Time
		if input.Body.DateOfBirth != nil {
			parsed, err := time.Parse("2006-01-02", *input.Body.DateOfBirth)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid date of birth format, use YYYY-MM-DD", err)
			}
			dateOfBirth = &parsed
		}

		schedules := make([]service.DoseScheduleInput, len(input.Body.DoseSchedules))
		for i, s := range input.Body.DoseSchedules {
			schedules[i] = service.DoseScheduleInput{Name: s.Name, Time: s.Time}
		}

		result, err := svc.CreateProfile(ctx, userID, input.Body.Name, dateOfBirth, input.Body.Timezone, schedules)
		if err != nil {
			if errors.Is(err, service.ErrInvalidTimezone) {
				return nil, huma.Error400BadRequest("Invalid timezone", err)
			}
			return nil, huma.Error500InternalServerError("Failed to create profile", err)
		}

		resp := &dto.CreateProfileOutput{}
		resp.Body.Profile = *result
		return resp, nil
	}
}

func ListProfilesHandler(svc service.ProfileService, tokenSvc auth.TokenValidator) func(context.Context, *dto.ListProfilesInput) (*dto.ListProfilesOutput, error) {
	return func(ctx context.Context, input *dto.ListProfilesInput) (*dto.ListProfilesOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		results, err := svc.ListProfiles(ctx, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to list profiles", err)
		}

		resp := &dto.ListProfilesOutput{}
		resp.Body.Profiles = results
		return resp, nil
	}
}

func GetProfileHandler(svc service.ProfileService, tokenSvc auth.TokenValidator) func(context.Context, *dto.GetProfileInput) (*dto.GetProfileOutput, error) {
	return func(ctx context.Context, input *dto.GetProfileInput) (*dto.GetProfileOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		result, err := svc.GetProfile(ctx, profileID, userID)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to get profile", err)
		}

		resp := &dto.GetProfileOutput{}
		resp.Body.Profile = *result
		return resp, nil
	}
}

func UpdateProfileHandler(svc service.ProfileService, tokenSvc auth.TokenValidator) func(context.Context, *dto.UpdateProfileInput) (*dto.UpdateProfileOutput, error) {
	return func(ctx context.Context, input *dto.UpdateProfileInput) (*dto.UpdateProfileOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		var dateOfBirth *time.Time
		if input.Body.DateOfBirth != nil {
			parsed, err := time.Parse("2006-01-02", *input.Body.DateOfBirth)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid date of birth format, use YYYY-MM-DD", err)
			}
			dateOfBirth = &parsed
		}

		result, err := svc.UpdateProfile(ctx, profileID, userID, input.Body.Name, dateOfBirth, input.Body.Timezone)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to profile", err)
			}
			if errors.Is(err, service.ErrInvalidTimezone) {
				return nil, huma.Error400BadRequest("Invalid timezone", err)
			}
			return nil, huma.Error500InternalServerError("Failed to update profile", err)
		}

		resp := &dto.UpdateProfileOutput{}
		resp.Body.Profile = *result
		return resp, nil
	}
}

func DeleteProfileHandler(svc service.ProfileService, tokenSvc auth.TokenValidator) func(context.Context, *dto.DeleteProfileInput) (*dto.DeleteProfileOutput, error) {
	return func(ctx context.Context, input *dto.DeleteProfileInput) (*dto.DeleteProfileOutput, error) {
		userID, err := auth.ExtractUserID(input.Authorization, tokenSvc)
		if err != nil {
			return nil, huma.Error401Unauthorized("Invalid or expired access token", err)
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		err = svc.DeleteProfile(ctx, profileID, userID)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to delete profile", err)
		}

		resp := &dto.DeleteProfileOutput{}
		resp.Body.Message = "Profile deleted successfully"
		return resp, nil
	}
}
