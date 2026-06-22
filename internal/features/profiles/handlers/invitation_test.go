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

func TestShareProfileHandler_Success(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	sharedWithID := uuid.UUID{2}
	invitationID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ShareProfile", mock.Anything, profileID, userID, service.ShareInput{
		SharedWithUserID: sharedWithID,
		Permissions:      []string{"profile:read"},
		ExpiresInDays:    7,
	}).Return(&service.InvitationResult{
		Invitation: service.InvitationDTO{
			ID:               invitationID,
			ProfileID:        profileID,
			SharedWithUserID: sharedWithID,
			GrantedByUserID:  userID,
			Permissions:      []string{"profile:read"},
			Status:           "pending",
			CreatedAt:        now,
		},
	}, nil)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.SharedWithUserID = sharedWithID
	input.Body.Permissions = []string{"profile:read"}
	input.Body.ExpiresInDays = 7

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pending", resp.Body.Invitation.Status)
	assert.Equal(t, []string{"profile:read"}, resp.Body.Invitation.Permissions)
}

func TestShareProfileHandler_InvalidToken(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "invalid-token").Return(nil, assert.AnError)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer invalid-token"
	input.ID = uuid.New().String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid or expired access token")
}

func TestShareProfileHandler_InvalidProfileID(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = "not-a-uuid"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid profile ID")
}

func TestShareProfileHandler_ProfileNotFound(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ShareProfile", mock.Anything, profileID, userID, mock.Anything).
		Return(nil, service.ErrProfileNotFound)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.SharedWithUserID = uuid.UUID{2}
	input.Body.Permissions = []string{"profile:read"}
	input.Body.ExpiresInDays = 7

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Profile not found")
}

func TestShareProfileHandler_SelfSharing(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ShareProfile", mock.Anything, profileID, userID, mock.Anything).
		Return(nil, service.ErrCannotShareWithSelf)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.SharedWithUserID = userID
	input.Body.Permissions = []string{"profile:read"}
	input.Body.ExpiresInDays = 7

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Cannot share profile with yourself")
}

func TestShareProfileHandler_InvalidPermissions(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ShareProfile", mock.Anything, profileID, userID, mock.Anything).
		Return(nil, service.ErrInvalidPermissions)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.SharedWithUserID = uuid.UUID{2}
	input.Body.Permissions = []string{}

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid permissions")
}

func TestShareProfileHandler_UserNotFound(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ShareProfile", mock.Anything, profileID, userID, mock.Anything).
		Return(nil, service.ErrUserNotFound)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.SharedWithUserID = uuid.UUID{2}
	input.Body.Permissions = []string{"profile:read"}
	input.Body.ExpiresInDays = 7

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "User not found")
}

func TestShareProfileHandler_UserAlreadySharing(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ShareProfile", mock.Anything, profileID, userID, mock.Anything).
		Return(nil, service.ErrUserAlreadySharing)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.SharedWithUserID = uuid.UUID{2}
	input.Body.Permissions = []string{"profile:read"}
	input.Body.ExpiresInDays = 7

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "User already has access")
}

func TestShareProfileHandler_InternalError(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ShareProfile", mock.Anything, profileID, userID, mock.Anything).
		Return(nil, assert.AnError)

	handler := handlers.ShareProfileHandler(mockSvc, mockTokenSvc)

	input := &dto.ShareProfileInput{}
	input.Authorization = "Bearer valid-token"
	input.ID = profileID.String()
	input.Body.SharedWithUserID = uuid.UUID{2}
	input.Body.Permissions = []string{"profile:read"}
	input.Body.ExpiresInDays = 7

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Failed to share profile")
}

func TestListInvitationsHandler_Success(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	profileID := uuid.New()
	invitationID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ListInvitations", mock.Anything, userID).Return([]service.InvitationResult{
		{
			Invitation: service.InvitationDTO{
				ID:               invitationID,
				ProfileID:        profileID,
				ProfileName:      "Test Profile",
				SharedWithUserID: userID,
				GrantedByUserID:  uuid.UUID{2},
				Permissions:      []string{"profile:read"},
				Status:           "pending",
				ExpiresAt:        timePtr(time.Now().Add(24 * time.Hour)),
				CreatedAt:        now,
			},
		},
	}, nil)

	handler := handlers.ListInvitationsHandler(mockSvc, mockTokenSvc)

	input := &dto.ListInvitationsInput{}
	input.Authorization = "Bearer valid-token"

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Body.Invitations, 1)
	assert.Equal(t, "pending", resp.Body.Invitations[0].Status)
	assert.Equal(t, "Test Profile", resp.Body.Invitations[0].ProfileName)
}

func TestListInvitationsHandler_InvalidToken(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "invalid-token").Return(nil, assert.AnError)

	handler := handlers.ListInvitationsHandler(mockSvc, mockTokenSvc)

	input := &dto.ListInvitationsInput{}
	input.Authorization = "Bearer invalid-token"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestListInvitationsHandler_InternalError(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("ListInvitations", mock.Anything, userID).Return(nil, assert.AnError)

	handler := handlers.ListInvitationsHandler(mockSvc, mockTokenSvc)

	input := &dto.ListInvitationsInput{}
	input.Authorization = "Bearer valid-token"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Failed to list invitations")
}

func TestAcceptInvitationHandler_Success(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	invitationID := uuid.New()
	profileID := uuid.New()
	now := time.Now()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("AcceptInvitation", mock.Anything, invitationID, userID).Return(&service.AcceptedProfileResult{
		Profile: dto.ProfileDTO{
			ID:        profileID,
			Name:      "Test Profile",
			Timezone:  "UTC",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Permissions: []string{"profile:read"},
	}, nil)

	handler := handlers.AcceptInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.AcceptInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = invitationID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test Profile", resp.Body.Profile.Name)
	assert.Equal(t, []string{"profile:read"}, resp.Body.Permissions)
}

func TestAcceptInvitationHandler_InvalidToken(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "invalid-token").Return(nil, assert.AnError)

	handler := handlers.AcceptInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.AcceptInvitationInput{}
	input.Authorization = "Bearer invalid-token"
	input.InvitationID = uuid.New().String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAcceptInvitationHandler_InvalidInvitationID(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	handler := handlers.AcceptInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.AcceptInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = "not-a-uuid"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid invitation ID")
}

func TestAcceptInvitationHandler_NotFound(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	invitationID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("AcceptInvitation", mock.Anything, invitationID, userID).Return(nil, service.ErrInvitationNotFound)

	handler := handlers.AcceptInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.AcceptInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = invitationID.String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invitation not found")
}

func TestAcceptInvitationHandler_Expired(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	invitationID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("AcceptInvitation", mock.Anything, invitationID, userID).Return(nil, service.ErrInvitationExpired)

	handler := handlers.AcceptInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.AcceptInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = invitationID.String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invitation has expired")
}

func TestAcceptInvitationHandler_AlreadyProcessed(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	invitationID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("AcceptInvitation", mock.Anything, invitationID, userID).Return(nil, service.ErrInvitationAlreadyProcessed)

	handler := handlers.AcceptInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.AcceptInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = invitationID.String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "already been processed")
}

func TestDeclineInvitationHandler_Success(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	invitationID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("DeclineInvitation", mock.Anything, invitationID, userID).Return(nil)

	handler := handlers.DeclineInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.DeclineInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = invitationID.String()

	resp, err := handler(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Invitation declined successfully", resp.Body.Message)
}

func TestDeclineInvitationHandler_InvalidToken(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	mockTokenSvc.On("ValidateAccessToken", "invalid-token").Return(nil, assert.AnError)

	handler := handlers.DeclineInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.DeclineInvitationInput{}
	input.Authorization = "Bearer invalid-token"
	input.InvitationID = uuid.New().String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestDeclineInvitationHandler_InvalidInvitationID(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	handler := handlers.DeclineInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.DeclineInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = "not-a-uuid"

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invalid invitation ID")
}

func TestDeclineInvitationHandler_NotFound(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	invitationID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("DeclineInvitation", mock.Anything, invitationID, userID).Return(service.ErrInvitationNotFound)

	handler := handlers.DeclineInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.DeclineInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = invitationID.String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Invitation not found")
}

func TestDeclineInvitationHandler_AlreadyProcessed(t *testing.T) {
	mockSvc := new(MockInvitationService)
	mockTokenSvc := new(MockTokenService)

	userID := uuid.UUID{1}
	invitationID := uuid.New()

	mockTokenSvc.On("ValidateAccessToken", "valid-token").Return(jwt.MapClaims{
		"sub": userID.String(),
	}, nil)

	mockSvc.On("DeclineInvitation", mock.Anything, invitationID, userID).Return(service.ErrInvitationAlreadyProcessed)

	handler := handlers.DeclineInvitationHandler(mockSvc, mockTokenSvc)

	input := &dto.DeclineInvitationInput{}
	input.Authorization = "Bearer valid-token"
	input.InvitationID = invitationID.String()

	resp, err := handler(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "already been processed")
}

func timePtr(t time.Time) *time.Time {
	return &t
}
