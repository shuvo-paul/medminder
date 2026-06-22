package handlers_test

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/features/profiles/dto"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/stretchr/testify/mock"
)

type MockProfileService struct {
	mock.Mock
}

func (m *MockProfileService) CreateProfile(ctx context.Context, userID uuid.UUID, name string, dateOfBirth *time.Time, timezone string, schedules []service.DoseScheduleInput) (*dto.ProfileDTO, error) {
	args := m.Called(ctx, userID, name, dateOfBirth, timezone, schedules)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileDTO), args.Error(1)
}

func (m *MockProfileService) GetProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) (*dto.ProfileDTO, error) {
	args := m.Called(ctx, profileID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileDTO), args.Error(1)
}

func (m *MockProfileService) ListProfiles(ctx context.Context, userID uuid.UUID) ([]dto.ProfileDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.ProfileDTO), args.Error(1)
}

func (m *MockProfileService) UpdateProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name *string, dateOfBirth *time.Time, timezone *string) (*dto.ProfileDTO, error) {
	args := m.Called(ctx, profileID, userID, name, dateOfBirth, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileDTO), args.Error(1)
}

func (m *MockProfileService) DeleteProfile(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, profileID, userID)
	return args.Error(0)
}

func (m *MockProfileService) HandleAccountDeletion(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockDoseScheduleService struct {
	mock.Mock
}

func (m *MockDoseScheduleService) CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, name string, timeStr string) (*dto.DoseScheduleDTO, error) {
	args := m.Called(ctx, profileID, userID, name, timeStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DoseScheduleDTO), args.Error(1)
}

func (m *MockDoseScheduleService) GetDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID) (*dto.DoseScheduleDTO, error) {
	args := m.Called(ctx, profileID, scheduleID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DoseScheduleDTO), args.Error(1)
}

func (m *MockDoseScheduleService) ListDoseSchedules(ctx context.Context, profileID uuid.UUID, userID uuid.UUID) ([]dto.DoseScheduleDTO, error) {
	args := m.Called(ctx, profileID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.DoseScheduleDTO), args.Error(1)
}

func (m *MockDoseScheduleService) UpdateDoseSchedule(ctx context.Context, profileID uuid.UUID, scheduleID uuid.UUID, userID uuid.UUID, name *string, timeStr *string) (*dto.DoseScheduleDTO, error) {
	args := m.Called(ctx, profileID, scheduleID, userID, name, timeStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DoseScheduleDTO), args.Error(1)
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

type MockInvitationService struct {
	mock.Mock
}

func (m *MockInvitationService) ShareProfile(ctx context.Context, profileID uuid.UUID, grantedByUserID uuid.UUID, input service.ShareInput) (*service.InvitationResult, error) {
	args := m.Called(ctx, profileID, grantedByUserID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.InvitationResult), args.Error(1)
}

func (m *MockInvitationService) ListInvitations(ctx context.Context, userID uuid.UUID) ([]service.InvitationResult, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.InvitationResult), args.Error(1)
}

func (m *MockInvitationService) AcceptInvitation(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) (*service.AcceptedProfileResult, error) {
	args := m.Called(ctx, invitationID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AcceptedProfileResult), args.Error(1)
}

func (m *MockInvitationService) DeclineInvitation(ctx context.Context, invitationID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, invitationID, userID)
	return args.Error(0)
}

type MockOwnershipTransferService struct {
	mock.Mock
}

func (m *MockOwnershipTransferService) InitiateTransfer(ctx context.Context, profileID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID) (*service.OwnershipTransferResult, error) {
	args := m.Called(ctx, profileID, fromUserID, toUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.OwnershipTransferResult), args.Error(1)
}

func (m *MockOwnershipTransferService) ListPendingTransfers(ctx context.Context, userID uuid.UUID) ([]service.OwnershipTransferResult, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.OwnershipTransferResult), args.Error(1)
}

func (m *MockOwnershipTransferService) AcceptTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, transferID, userID)
	return args.Error(0)
}

func (m *MockOwnershipTransferService) DeclineTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, transferID, userID)
	return args.Error(0)
}

func (m *MockOwnershipTransferService) CancelTransfer(ctx context.Context, transferID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, transferID, userID)
	return args.Error(0)
}
