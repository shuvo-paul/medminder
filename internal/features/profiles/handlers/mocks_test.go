package handlers_test

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/stretchr/testify/mock"
)

type MockProfileService struct {
	mock.Mock
}

func (m *MockProfileService) CreateProfile(ctx context.Context, userID uuid.UUID, name string, dateOfBirth *time.Time, timezone string, schedules []service.DoseScheduleInput) (*service.ProfileResult, error) {
	args := m.Called(ctx, userID, name, dateOfBirth, timezone, schedules)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ProfileResult), args.Error(1)
}

func (m *MockProfileService) GetProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) (*service.ProfileResult, error) {
	args := m.Called(ctx, profileID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ProfileResult), args.Error(1)
}

func (m *MockProfileService) ListProfiles(ctx context.Context, userID uuid.UUID) ([]service.ProfileResult, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.ProfileResult), args.Error(1)
}

func (m *MockProfileService) UpdateProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name *string, dateOfBirth *time.Time, timezone *string) (*service.ProfileResult, error) {
	args := m.Called(ctx, profileID, userID, name, dateOfBirth, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ProfileResult), args.Error(1)
}

func (m *MockProfileService) DeleteProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, profileID, userID)
	return args.Error(0)
}

type MockDoseScheduleService struct {
	mock.Mock
}

func (m *MockDoseScheduleService) CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name string, timeStr string) (*service.DoseScheduleResult, error) {
	args := m.Called(ctx, profileID, userID, name, timeStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DoseScheduleResult), args.Error(1)
}

func (m *MockDoseScheduleService) GetDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) (*service.DoseScheduleResult, error) {
	args := m.Called(ctx, profileID, scheduleID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DoseScheduleResult), args.Error(1)
}

func (m *MockDoseScheduleService) ListDoseSchedules(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) ([]service.DoseScheduleResult, error) {
	args := m.Called(ctx, profileID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.DoseScheduleResult), args.Error(1)
}

func (m *MockDoseScheduleService) UpdateDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID, name *string, timeStr *string) (*service.DoseScheduleResult, error) {
	args := m.Called(ctx, profileID, scheduleID, userID, name, timeStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DoseScheduleResult), args.Error(1)
}

func (m *MockDoseScheduleService) DeleteDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, profileID, scheduleID, userID)
	return args.Error(0)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

func (m *MockTokenService) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GenerateRefreshToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) HashRefreshToken(token string) string {
	args := m.Called(token)
	return args.String(0)
}
