package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/dto"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/handlers"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/service"
	"github.com/shuvo-paul/medminder/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGuestAccessService struct {
	mock.Mock
}

func (m *MockGuestAccessService) CreateToken(ctx context.Context, profileID uuid.UUID, label string, permissions []string, expiresInDays int) (*service.TokenResult, error) {
	args := m.Called(ctx, profileID, label, permissions, expiresInDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenResult), args.Error(1)
}

func (m *MockGuestAccessService) ListTokens(ctx context.Context, profileID uuid.UUID) ([]service.TokenResult, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.TokenResult), args.Error(1)
}

func (m *MockGuestAccessService) RevokeToken(ctx context.Context, tokenID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, tokenID, userID)
	return args.Error(0)
}

func (m *MockGuestAccessService) Authenticate(ctx context.Context, rawToken string) (*service.AuthenticatedToken, error) {
	args := m.Called(ctx, rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthenticatedToken), args.Error(1)
}

type MockDoseScheduleRepository struct {
	mock.Mock
}

func (m *MockDoseScheduleRepository) ListDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) ([]db.DoseSchedule, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.DoseSchedule), args.Error(1)
}

func (m *MockDoseScheduleRepository) GetDoseScheduleByID(ctx context.Context, id uuid.UUID) (db.DoseSchedule, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.DoseSchedule), args.Error(1)
}

func (m *MockDoseScheduleRepository) CreateDoseSchedule(ctx context.Context, profileID uuid.UUID, name string, t time.Time) (db.DoseSchedule, error) {
	args := m.Called(ctx, profileID, name, t)
	return args.Get(0).(db.DoseSchedule), args.Error(1)
}

func (m *MockDoseScheduleRepository) UpdateDoseSchedule(ctx context.Context, id uuid.UUID, name string, t time.Time) (db.DoseSchedule, error) {
	args := m.Called(ctx, id, name, t)
	return args.Get(0).(db.DoseSchedule), args.Error(1)
}

func (m *MockDoseScheduleRepository) DeleteDoseSchedule(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDoseScheduleRepository) DeleteDoseSchedulesByProfile(ctx context.Context, profileID uuid.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func TestCreateGuestAccessHandler_Success(t *testing.T) {
	mockSvc := new(MockGuestAccessService)

	profileID := uuid.New()
	tokenID := uuid.New()
	expiresAt := time.Now().AddDate(0, 0, 30)

	mockSvc.On("CreateToken", mock.Anything, profileID, "", mock.Anything, 30).Return(&service.TokenResult{
		ID:        tokenID,
		RawToken:  "abc123",
		ExpiresAt: expiresAt,
	}, nil)

	handler := handlers.CreateGuestAccessHandler(mockSvc)

	input := &dto.CreateGuestAccessInput{}
	input.ID = profileID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, tokenID, resp.Body.ID)
	assert.Equal(t, "abc123", resp.Body.Token)
}

func TestListGuestAccessTokensHandler_Success(t *testing.T) {
	mockSvc := new(MockGuestAccessService)

	profileID := uuid.New()
	now := time.Now()

	mockSvc.On("ListTokens", mock.Anything, profileID).Return([]service.TokenResult{
		{ID: uuid.New(), Label: "token-1", ExpiresAt: now.AddDate(0, 0, 30), CreatedAt: now},
		{ID: uuid.New(), Label: "token-2", ExpiresAt: now.AddDate(0, 0, 7), CreatedAt: now},
	}, nil)

	handler := handlers.ListGuestAccessTokensHandler(mockSvc)

	input := &dto.ListGuestAccessTokensInput{}
	input.ID = profileID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.Len(t, resp.Body.Tokens, 2)
}

func TestRevokeGuestAccessHandler_Success(t *testing.T) {
	mockSvc := new(MockGuestAccessService)

	userID := uuid.UUID{1}
	tokenID := uuid.New()

	mockSvc.On("RevokeToken", mock.Anything, tokenID, userID).Return(nil)

	handler := handlers.RevokeGuestAccessHandler(mockSvc)

	input := &dto.RevokeGuestAccessInput{}
	input.TokenID = tokenID.String()

	ctx := middleware.ContextWithUserID(context.Background(), userID)
	resp, err := handler(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, "Guest access token revoked successfully", resp.Body.Message)
}

func TestRevokeGuestAccessHandler_NotFound(t *testing.T) {
	mockSvc := new(MockGuestAccessService)

	userID := uuid.UUID{1}
	tokenID := uuid.New()

	mockSvc.On("RevokeToken", mock.Anything, tokenID, userID).Return(service.ErrGuestTokenNotFound)

	handler := handlers.RevokeGuestAccessHandler(mockSvc)

	input := &dto.RevokeGuestAccessInput{}
	input.TokenID = tokenID.String()

	ctx := middleware.ContextWithUserID(context.Background(), userID)
	_, err := handler(ctx, input)

	assert.Error(t, err)
}

func TestRevokeGuestAccessHandler_MissingUserID(t *testing.T) {
	mockSvc := new(MockGuestAccessService)

	handler := handlers.RevokeGuestAccessHandler(mockSvc)

	input := &dto.RevokeGuestAccessInput{}
	input.TokenID = uuid.New().String()

	_, err := handler(context.Background(), input)

	assert.Error(t, err)
}

func TestGuestListMedicationsHandler_Success(t *testing.T) {
	mockAuthSvc := new(MockGuestAccessService)
	mockScheduleRepo := new(MockDoseScheduleRepository)

	profileID := uuid.New()
	now := time.Now()

	mockAuthSvc.On("Authenticate", mock.Anything, "valid-token").Return(&service.AuthenticatedToken{
		TokenID:     uuid.New(),
		ProfileID:   profileID,
		Permissions: []string{"medication:read", "reminder:read"},
	}, nil)

	mockScheduleRepo.On("ListDoseSchedulesByProfile", mock.Anything, profileID).Return([]db.DoseSchedule{
		{ID: uuid.New(), ProfileID: profileID, Name: "Aspirin", Time: now, CreatedAt: now, UpdatedAt: now},
	}, nil)

	handler := handlers.GuestListMedicationsHandler(mockAuthSvc, mockScheduleRepo)

	input := &dto.GuestMedicationInput{}
	input.Token = "valid-token"

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.Len(t, resp.Body.Medications, 1)
	assert.Equal(t, "Aspirin", resp.Body.Medications[0].Name)
}

func TestGuestListMedicationsHandler_ExpiredToken(t *testing.T) {
	mockAuthSvc := new(MockGuestAccessService)
	mockScheduleRepo := new(MockDoseScheduleRepository)

	mockAuthSvc.On("Authenticate", mock.Anything, "expired-token").Return(nil, service.ErrGuestTokenExpired)

	handler := handlers.GuestListMedicationsHandler(mockAuthSvc, mockScheduleRepo)

	input := &dto.GuestMedicationInput{}
	input.Token = "expired-token"

	_, err := handler(context.Background(), input)

	assert.Error(t, err)
}

func TestGuestListMedicationsHandler_NoPermission(t *testing.T) {
	mockAuthSvc := new(MockGuestAccessService)
	mockScheduleRepo := new(MockDoseScheduleRepository)

	profileID := uuid.New()

	mockAuthSvc.On("Authenticate", mock.Anything, "no-perm-token").Return(&service.AuthenticatedToken{
		TokenID:     uuid.New(),
		ProfileID:   profileID,
		Permissions: []string{"reminder:read"},
	}, nil)

	handler := handlers.GuestListMedicationsHandler(mockAuthSvc, mockScheduleRepo)

	input := &dto.GuestMedicationInput{}
	input.Token = "no-perm-token"

	_, err := handler(context.Background(), input)

	assert.Error(t, err)
}

func TestGuestGetReminderHandler_Success(t *testing.T) {
	mockAuthSvc := new(MockGuestAccessService)
	mockScheduleRepo := new(MockDoseScheduleRepository)

	profileID := uuid.New()
	reminderID := uuid.New()
	now := time.Now()

	mockAuthSvc.On("Authenticate", mock.Anything, "valid-token").Return(&service.AuthenticatedToken{
		TokenID:     uuid.New(),
		ProfileID:   profileID,
		Permissions: []string{"medication:read", "reminder:read"},
	}, nil)

	mockScheduleRepo.On("GetDoseScheduleByID", mock.Anything, reminderID).Return(db.DoseSchedule{
		ID: reminderID, ProfileID: profileID, Name: "Evening pill", Time: now, CreatedAt: now, UpdatedAt: now,
	}, nil)

	handler := handlers.GuestGetReminderHandler(mockAuthSvc, mockScheduleRepo)

	input := &dto.GuestReminderInput{}
	input.Token = "valid-token"
	input.ReminderID = reminderID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, "Evening pill", resp.Body.Reminder.Name)
}

func TestGuestGetReminderHandler_NotFound(t *testing.T) {
	mockAuthSvc := new(MockGuestAccessService)
	mockScheduleRepo := new(MockDoseScheduleRepository)

	profileID := uuid.New()
	reminderID := uuid.New()

	mockAuthSvc.On("Authenticate", mock.Anything, "valid-token").Return(&service.AuthenticatedToken{
		TokenID:     uuid.New(),
		ProfileID:   profileID,
		Permissions: []string{"medication:read", "reminder:read"},
	}, nil)

	mockScheduleRepo.On("GetDoseScheduleByID", mock.Anything, reminderID).Return(db.DoseSchedule{}, sql.ErrNoRows)

	handler := handlers.GuestGetReminderHandler(mockAuthSvc, mockScheduleRepo)

	input := &dto.GuestReminderInput{}
	input.Token = "valid-token"
	input.ReminderID = reminderID.String()

	_, err := handler(context.Background(), input)

	assert.Error(t, err)
}

func TestGuestGetReminderHandler_WrongProfile(t *testing.T) {
	mockAuthSvc := new(MockGuestAccessService)
	mockScheduleRepo := new(MockDoseScheduleRepository)

	profileID := uuid.New()
	reminderID := uuid.New()
	now := time.Now()

	mockAuthSvc.On("Authenticate", mock.Anything, "valid-token").Return(&service.AuthenticatedToken{
		TokenID:     uuid.New(),
		ProfileID:   profileID,
		Permissions: []string{"medication:read", "reminder:read"},
	}, nil)

	otherProfileID := uuid.New()
	mockScheduleRepo.On("GetDoseScheduleByID", mock.Anything, reminderID).Return(db.DoseSchedule{
		ID: reminderID, ProfileID: otherProfileID, Name: "Other's pill", Time: now, CreatedAt: now, UpdatedAt: now,
	}, nil)

	handler := handlers.GuestGetReminderHandler(mockAuthSvc, mockScheduleRepo)

	input := &dto.GuestReminderInput{}
	input.Token = "valid-token"
	input.ReminderID = reminderID.String()

	_, err := handler(context.Background(), input)

	assert.Error(t, err)
}
