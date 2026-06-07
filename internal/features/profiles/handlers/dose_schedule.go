package handlers

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
)

func ListSchedulesHandler(svc service.DoseScheduleService, profileSvc service.ProfileService, tokenSvc TokenServiceInterface) func(context.Context, *dto.ListSchedulesInput) (*dto.ListSchedulesOutput, error) {
	return func(ctx context.Context, input *dto.ListSchedulesInput) (*dto.ListSchedulesOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		_, err = profileSvc.GetProfile(ctx, profileID, userID)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to get profile", err)
		}

		results, err := svc.ListDoseSchedules(ctx, profileID, userID)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to list schedules", err)
		}

		resp := &dto.ListSchedulesOutput{}
		resp.Body.Schedules = make([]dto.DoseScheduleDTO, len(results))
		for i, r := range results {
			resp.Body.Schedules[i] = toDoseScheduleDTO(r.Schedule)
		}
		return resp, nil
	}
}

func CreateScheduleHandler(svc service.DoseScheduleService, profileSvc service.ProfileService, tokenSvc TokenServiceInterface) func(context.Context, *dto.CreateScheduleInput) (*dto.CreateScheduleOutput, error) {
	return func(ctx context.Context, input *dto.CreateScheduleInput) (*dto.CreateScheduleOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		_, err = profileSvc.GetProfile(ctx, profileID, userID)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to get profile", err)
		}

		result, err := svc.CreateDoseSchedule(ctx, profileID, userID, input.Body.Name, input.Body.Time)
		if err != nil {
			if errors.Is(err, service.ErrInvalidTimezone) {
				return nil, huma.Error400BadRequest("Invalid time format, use HH:MM", err)
			}
			return nil, huma.Error500InternalServerError("Failed to create schedule", err)
		}

		resp := &dto.CreateScheduleOutput{}
		resp.Body.Schedule = toDoseScheduleDTO(result.Schedule)
		return resp, nil
	}
}

func GetScheduleHandler(svc service.DoseScheduleService, profileSvc service.ProfileService, tokenSvc TokenServiceInterface) func(context.Context, *dto.GetScheduleInput) (*dto.GetScheduleOutput, error) {
	return func(ctx context.Context, input *dto.GetScheduleInput) (*dto.GetScheduleOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		scheduleID, err := uuid.Parse(input.ScheduleID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid schedule ID", err)
		}

		_, err = profileSvc.GetProfile(ctx, profileID, userID)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to get profile", err)
		}

		result, err := svc.GetDoseSchedule(ctx, profileID, scheduleID, userID)
		if err != nil {
			if errors.Is(err, service.ErrScheduleNotFound) {
				return nil, huma.Error404NotFound("Schedule not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to schedule", err)
			}
			return nil, huma.Error500InternalServerError("Failed to get schedule", err)
		}

		resp := &dto.GetScheduleOutput{}
		resp.Body.Schedule = toDoseScheduleDTO(result.Schedule)
		return resp, nil
	}
}

func UpdateScheduleHandler(svc service.DoseScheduleService, profileSvc service.ProfileService, tokenSvc TokenServiceInterface) func(context.Context, *dto.UpdateScheduleInput) (*dto.UpdateScheduleOutput, error) {
	return func(ctx context.Context, input *dto.UpdateScheduleInput) (*dto.UpdateScheduleOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		scheduleID, err := uuid.Parse(input.ScheduleID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid schedule ID", err)
		}

		_, err = profileSvc.GetProfile(ctx, profileID, userID)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to get profile", err)
		}

		result, err := svc.UpdateDoseSchedule(ctx, profileID, scheduleID, userID, input.Body.Name, input.Body.Time)
		if err != nil {
			if errors.Is(err, service.ErrScheduleNotFound) {
				return nil, huma.Error404NotFound("Schedule not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to schedule", err)
			}
			if errors.Is(err, service.ErrInvalidTimezone) {
				return nil, huma.Error400BadRequest("Invalid time format, use HH:MM", err)
			}
			return nil, huma.Error500InternalServerError("Failed to update schedule", err)
		}

		resp := &dto.UpdateScheduleOutput{}
		resp.Body.Schedule = toDoseScheduleDTO(result.Schedule)
		return resp, nil
	}
}

func DeleteScheduleHandler(svc service.DoseScheduleService, profileSvc service.ProfileService, tokenSvc TokenServiceInterface) func(context.Context, *dto.DeleteScheduleInput) (*dto.DeleteScheduleOutput, error) {
	return func(ctx context.Context, input *dto.DeleteScheduleInput) (*dto.DeleteScheduleOutput, error) {
		userID, err := ExtractUserIDFromAuth(input.Authorization, tokenSvc)
		if err != nil {
			return nil, err
		}

		profileID, err := uuid.Parse(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid profile ID", err)
		}

		scheduleID, err := uuid.Parse(input.ScheduleID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid schedule ID", err)
		}

		_, err = profileSvc.GetProfile(ctx, profileID, userID)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				return nil, huma.Error404NotFound("Profile not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to profile", err)
			}
			return nil, huma.Error500InternalServerError("Failed to get profile", err)
		}

		err = svc.DeleteDoseSchedule(ctx, profileID, scheduleID, userID)
		if err != nil {
			if errors.Is(err, service.ErrScheduleNotFound) {
				return nil, huma.Error404NotFound("Schedule not found", err)
			}
			if errors.Is(err, service.ErrUnauthorizedAccess) {
				return nil, huma.Error403Forbidden("Unauthorized access to schedule", err)
			}
			return nil, huma.Error500InternalServerError("Failed to delete schedule", err)
		}

		resp := &dto.DeleteScheduleOutput{}
		resp.Body.Message = "Schedule deleted successfully"
		return resp, nil
	}
}
