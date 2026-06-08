package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/handlers"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateProfile_Success(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	schedules := []service.DoseScheduleInput{
		{Name: "Morning", Time: "08:00"},
	}

	mockSvc.On("CreateProfile", mock.Anything, "Test Profile", mock.Anything, "America/New_York", schedules).Return(&service.ProfileResult{
		Profile: service.ProfileDTO{
			ID:          profileID,
			Name:        "Test Profile",
			DateOfBirth: func() *string { s := "1990-01-15"; return &s }(),
			Timezone:    "America/New_York",
			CreatedAt:   now,
			UpdatedAt:   now,
			Schedules: []service.DoseScheduleDTO{
				{ID: uuid.New(), ProfileID: profileID, Name: "Morning", Time: "08:00", CreatedAt: now, UpdatedAt: now},
			},
		},
	}, nil)

	handler := handlers.CreateProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.CreateProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.Body.Name = "Test Profile"
	input.Body.DateOfBirth = func() *string { s := "1990-01-15"; return &s }()
	input.Body.Timezone = "America/New_York"
	input.Body.DoseSchedules = []dto.DoseScheduleInput{
		{Name: "Morning", Time: "08:00"},
	}

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test Profile", resp.Body.Profile.Name)
	assert.Equal(t, "America/New_York", resp.Body.Profile.Timezone)
}

func TestCreateProfile_InvalidToken(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "invalid-token").Return(nil, assert.AnError)

	handler := handlers.CreateProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.CreateProfileInput{}
	input.Authorization = "Bearer invalid-token"
	input.Body.Name = "Test Profile"
	input.Body.Timezone = "UTC"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired access token")
}

func TestCreateProfile_InvalidTimezone(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("CreateProfile", mock.Anything, "Test Profile", mock.Anything, "Invalid/Timezone", mock.Anything).Return(nil, service.ErrInvalidTimezone)

	handler := handlers.CreateProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.CreateProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.Body.Name = "Test Profile"
	input.Body.Timezone = "Invalid/Timezone"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid timezone")
}

func TestListProfiles_Success(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ListProfiles", mock.Anything, userID).Return([]service.ProfileResult{
		{
			Profile: service.ProfileDTO{
				ID:        profileID,
				Name:      "Profile 1",
				Timezone:  "UTC",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}, nil)

	handler := handlers.ListProfilesHandler(mockSvc, mockTokenSvc)

	input := &dto.ListProfilesInput{}
	input.Authorization = "Bearer valid-token"

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Body.Profiles, 1)
	assert.Equal(t, "Profile 1", resp.Body.Profiles[0].Name)
}

func TestGetProfile_Success(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("GetProfile", mock.Anything, profileID, userID).Return(&service.ProfileResult{
		Profile: service.ProfileDTO{
			ID:        profileID,
			Name:      "Test Profile",
			Timezone:  "UTC",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil)

	handler := handlers.GetProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.GetProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test Profile", resp.Body.Profile.Name)
}

func TestGetProfile_NotFound(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("GetProfile", mock.Anything, profileID, userID).Return(nil, service.ErrProfileNotFound)

	handler := handlers.GetProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.GetProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Profile not found")
}

func TestUpdateProfile_Success(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	now := time.Now()

	newName := "Updated Profile"
	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("UpdateProfile", mock.Anything, profileID, userID, &newName, mock.Anything, (*string)(nil)).Return(&service.ProfileResult{
		Profile: service.ProfileDTO{
			ID:        profileID,
			Name:      "Updated Profile",
			Timezone:  "UTC",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil)

	handler := handlers.UpdateProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.UpdateProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.Name = &newName

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated Profile", resp.Body.Profile.Name)
}

func TestDeleteProfile_Success(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("DeleteProfile", mock.Anything, profileID, userID).Return(nil)

	handler := handlers.DeleteProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.DeleteProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Profile deleted successfully", resp.Body.Message)
}

func TestDeleteProfile_NotFound(t *testing.T) {
	mockSvc := new(MockProfileService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("DeleteProfile", mock.Anything, profileID, userID).Return(service.ErrProfileNotFound)

	handler := handlers.DeleteProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.DeleteProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}
