package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/dto"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/handlers"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/service"
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

type MockPermissionChecker struct {
	mock.Mock
}

func (m *MockPermissionChecker) HasAnyPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permissions []string) (bool, error) {
	args := m.Called(ctx, profileID, userID, permissions)
	return args.Bool(0), args.Error(1)
}

func TestCreateGuestAccessHandler_Success(t *testing.T) {
	mockSvc := new(MockGuestAccessService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	tokenID := uuid.New()
	expiresAt := time.Now().AddDate(0, 0, 30)

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("CreateToken", mock.Anything, profileID, "", mock.Anything, 30).Return(&service.TokenResult{
		ID:        tokenID,
		RawToken:  "abc123",
		ExpiresAt: expiresAt,
	}, nil)

	handler := handlers.CreateGuestAccessHandler(mockSvc, mockTokenSvc)

	input := &dto.CreateGuestAccessInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, tokenID, resp.Body.ID)
	assert.Equal(t, "abc123", resp.Body.Token)
}

func TestCreateGuestAccessHandler_InvalidToken(t *testing.T) {
	mockSvc := new(MockGuestAccessService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "invalid-token").Return(nil, assert.AnError)

	handler := handlers.CreateGuestAccessHandler(mockSvc, mockTokenSvc)

	input := &dto.CreateGuestAccessInput{}
	input.Authorization = "Bearer invalid-token"
	input.ID = uuid.New().String()

	_, err := handler(context.Background(), input)

	assert.Error(t, err)
}

func TestListGuestAccessTokensHandler_Success(t *testing.T) {
	mockSvc := new(MockGuestAccessService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ListTokens", mock.Anything, profileID).Return([]service.TokenResult{
		{ID: uuid.New(), Label: "token-1", ExpiresAt: now.AddDate(0, 0, 30), CreatedAt: now},
		{ID: uuid.New(), Label: "token-2", ExpiresAt: now.AddDate(0, 0, 7), CreatedAt: now},
	}, nil)

	handler := handlers.ListGuestAccessTokensHandler(mockSvc, mockTokenSvc)

	input := &dto.ListGuestAccessTokensInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.Len(t, resp.Body.Tokens, 2)
}

func TestRevokeGuestAccessHandler_Success(t *testing.T) {
	mockSvc := new(MockGuestAccessService)
	mockTokenSvc := new(MockTokenService)
	mockPermChecker := new(MockPermissionChecker)

	userID := uuid.UUID{1}
	tokenID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("RevokeToken", mock.Anything, tokenID, userID).Return(nil)

	handler := handlers.RevokeGuestAccessHandler(mockSvc, mockTokenSvc, mockPermChecker)

	input := &dto.RevokeGuestAccessInput{}
	input.Authorization = "Bearer valid-token"
	input.TokenID = tokenID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, "Guest access token revoked successfully", resp.Body.Message)
}

func TestRevokeGuestAccessHandler_NotFound(t *testing.T) {
	mockSvc := new(MockGuestAccessService)
	mockTokenSvc := new(MockTokenService)
	mockPermChecker := new(MockPermissionChecker)

	userID := uuid.UUID{1}
	tokenID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("RevokeToken", mock.Anything, tokenID, userID).Return(service.ErrGuestTokenNotFound)

	handler := handlers.RevokeGuestAccessHandler(mockSvc, mockTokenSvc, mockPermChecker)

	input := &dto.RevokeGuestAccessInput{}
	input.Authorization = "Bearer valid-token"
	input.TokenID = tokenID.String()

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
