package service_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

func makeExpiry(days int) time.Time {
	return time.Now().AddDate(0, 0, days)
}

func TestInvitationService_ShareProfile(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*MockProfileRepository, *MockDoseScheduleRepository, *MockPermissionChecker)
		input struct {
			profileID       uuid.UUID
			grantedByUserID uuid.UUID
			sharedWithID    uuid.UUID
			permissions     []string
			expiresInDays   int
		}
		expect func(*testing.T, *service.InvitationResult, error)
	}{
		{
			name: "ShareProfile_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}, Name: "Test Profile"}, nil)
				profileRepo.On("UserExists", mock.Anything, mock.Anything).Return(true, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, mock.Anything, mock.Anything).
					Return(db.ProfilePermission{}, sql.ErrNoRows)
				profileRepo.On("CreateInvitation", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(db.ProfilePermission{
						ID:               uuid.MustParse("00000000-0000-0000-0000-000000000001"),
						ProfileID:        uuid.UUID{1},
						SharedWithUserID: uuid.UUID{2},
						GrantedByUserID:  uuid.UUID{1},
						Permissions:      json.RawMessage(`["profile:read"]`),
						Status:           "pending",
						ExpiresAt:        makeNullTime(makeExpiry(7)),
						CreatedAt:        time.Now(),
					}, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{"profile:read"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "pending", result.Invitation.Status)
				assert.Equal(t, []string{"profile:read"}, result.Invitation.Permissions)
			},
		},
		{
			name: "ShareProfile_ProfileNotFound",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{}, sql.ErrNoRows)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{"profile:read"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.ErrorIs(t, err, service.ErrProfileNotFound)
				assert.Nil(t, result)
			},
		},
		{
			name: "ShareProfile_SelfSharing",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}}, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{1},
				permissions:     []string{"profile:read"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.ErrorIs(t, err, service.ErrCannotShareWithSelf)
				assert.Nil(t, result)
			},
		},
		{
			name: "ShareProfile_InvalidPermission_Owner",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}}, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{"profile:owner"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.ErrorIs(t, err, service.ErrInvalidPermissions)
				assert.Nil(t, result)
			},
		},
		{
			name: "ShareProfile_InvalidPermission_Unknown",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}}, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{"profile:delete"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.ErrorIs(t, err, service.ErrInvalidPermissions)
				assert.Nil(t, result)
			},
		},
		{
			name: "ShareProfile_EmptyPermissions",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}}, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.ErrorIs(t, err, service.ErrInvalidPermissions)
				assert.Nil(t, result)
			},
		},
		{
			name: "ShareProfile_UserNotFound",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}, Name: "Test"}, nil)
				profileRepo.On("UserExists", mock.Anything, mock.Anything).Return(false, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{"profile:read"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.ErrorIs(t, err, service.ErrUserNotFound)
				assert.Nil(t, result)
			},
		},
		{
			name: "ShareProfile_AlreadySharing_Accepted",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}, Name: "Test"}, nil)
				profileRepo.On("UserExists", mock.Anything, mock.Anything).Return(true, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, mock.Anything, mock.Anything).
					Return(db.ProfilePermission{
						ID:               uuid.MustParse("00000000-0000-0000-0000-000000000002"),
						ProfileID:        uuid.UUID{1},
						SharedWithUserID: uuid.UUID{2},
						Status:           "accepted",
					}, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{"profile:read"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.ErrorIs(t, err, service.ErrUserAlreadySharing)
				assert.Nil(t, result)
			},
		},
		{
			name: "ShareProfile_AlreadySharing_Pending",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}, Name: "Test"}, nil)
				profileRepo.On("UserExists", mock.Anything, mock.Anything).Return(true, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, mock.Anything, mock.Anything).
					Return(db.ProfilePermission{
						ID:               uuid.MustParse("00000000-0000-0000-0000-000000000002"),
						ProfileID:        uuid.UUID{1},
						SharedWithUserID: uuid.UUID{2},
						Status:           "pending",
					}, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{"profile:read"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.ErrorIs(t, err, service.ErrUserAlreadySharing)
				assert.Nil(t, result)
			},
		},
		{
			name: "ShareProfile_DeclinedCanReshare",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, mock.Anything).
					Return(db.Profile{ID: uuid.UUID{1}, Name: "Test Profile"}, nil)
				profileRepo.On("UserExists", mock.Anything, mock.Anything).Return(true, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, mock.Anything, mock.Anything).
					Return(db.ProfilePermission{
						ID:               uuid.MustParse("00000000-0000-0000-0000-000000000002"),
						ProfileID:        uuid.UUID{1},
						SharedWithUserID: uuid.UUID{2},
						Status:           "declined",
					}, nil)
				profileRepo.On("CreateInvitation", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(db.ProfilePermission{
						ID:               uuid.MustParse("00000000-0000-0000-0000-000000000003"),
						ProfileID:        uuid.UUID{1},
						SharedWithUserID: uuid.UUID{2},
						GrantedByUserID:  uuid.UUID{1},
						Permissions:      json.RawMessage(`["profile:read"]`),
						Status:           "pending",
						ExpiresAt:        makeNullTime(makeExpiry(7)),
						CreatedAt:        time.Now(),
					}, nil)
			},
			input: struct {
				profileID       uuid.UUID
				grantedByUserID uuid.UUID
				sharedWithID    uuid.UUID
				permissions     []string
				expiresInDays   int
			}{
				profileID:       uuid.UUID{1},
				grantedByUserID: uuid.UUID{1},
				sharedWithID:    uuid.UUID{2},
				permissions:     []string{"profile:read"},
				expiresInDays:   7,
			},
			expect: func(t *testing.T, result *service.InvitationResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "pending", result.Invitation.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			permChecker := new(MockPermissionChecker)
			svc := service.NewInvitationService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo, permChecker)
			result, err := svc.ShareProfile(context.Background(), tt.input.profileID, tt.input.grantedByUserID, service.ShareInput{
				SharedWithUserID: tt.input.sharedWithID,
				Permissions:      tt.input.permissions,
				ExpiresInDays:    tt.input.expiresInDays,
			})
			tt.expect(t, result, err)
		})
	}
}

func TestInvitationService_ListInvitations(t *testing.T) {
	profileID := uuid.UUID{1}
	userID := uuid.UUID{2}

	tests := []struct {
		name   string
		setup  func(*MockProfileRepository, *MockDoseScheduleRepository, *MockPermissionChecker)
		expect func(*testing.T, []service.InvitationResult, error)
	}{
		{
			name: "ListInvitations_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("ListProfilePermissionsByUser", mock.Anything, userID).
					Return([]db.ProfilePermission{
						{
							ID:               uuid.MustParse("00000000-0000-0000-0000-000000000001"),
							ProfileID:        profileID,
							SharedWithUserID: userID,
							GrantedByUserID:  uuid.UUID{1},
							Permissions:      json.RawMessage(`["profile:read"]`),
							Status:           "pending",
							ExpiresAt:        makeNullTime(makeExpiry(7)),
							CreatedAt:        time.Now(),
						},
					}, nil)
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile"}, nil)
			},
			expect: func(t *testing.T, results []service.InvitationResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "pending", results[0].Invitation.Status)
				assert.Equal(t, "Test Profile", results[0].Invitation.ProfileName)
			},
		},
		{
			name: "ListInvitations_FiltersAccepted",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("ListProfilePermissionsByUser", mock.Anything, userID).
					Return([]db.ProfilePermission{
						{
							ID:               uuid.MustParse("00000000-0000-0000-0000-000000000001"),
							ProfileID:        profileID,
							SharedWithUserID: userID,
							Permissions:      json.RawMessage(`["profile:read"]`),
							Status:           "accepted",
							CreatedAt:        time.Now(),
						},
					}, nil)
			},
			expect: func(t *testing.T, results []service.InvitationResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 0)
			},
		},
		{
			name: "ListInvitations_FiltersExpired",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("ListProfilePermissionsByUser", mock.Anything, userID).
					Return([]db.ProfilePermission{
						{
							ID:               uuid.MustParse("00000000-0000-0000-0000-000000000001"),
							ProfileID:        profileID,
							SharedWithUserID: userID,
							Permissions:      json.RawMessage(`["profile:read"]`),
							Status:           "pending",
							ExpiresAt:        makeNullTime(time.Now().AddDate(0, 0, -1)),
							CreatedAt:        time.Now(),
						},
					}, nil)
			},
			expect: func(t *testing.T, results []service.InvitationResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 0)
			},
		},
		{
			name: "ListInvitations_FiltersExpiredAndReturnsPending",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("ListProfilePermissionsByUser", mock.Anything, userID).
					Return([]db.ProfilePermission{
						{
							ID:               uuid.MustParse("00000000-0000-0000-0000-000000000001"),
							ProfileID:        profileID,
							SharedWithUserID: userID,
							Permissions:      json.RawMessage(`["profile:read"]`),
							Status:           "pending",
							ExpiresAt:        makeNullTime(time.Now().AddDate(0, 0, -1)),
							CreatedAt:        time.Now(),
						},
						{
							ID:               uuid.MustParse("00000000-0000-0000-0000-000000000002"),
							ProfileID:        profileID,
							SharedWithUserID: userID,
							Permissions:      json.RawMessage(`["profile:write"]`),
							Status:           "pending",
							ExpiresAt:        makeNullTime(makeExpiry(7)),
							CreatedAt:        time.Now(),
						},
					}, nil)
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile"}, nil)
			},
			expect: func(t *testing.T, results []service.InvitationResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, []string{"profile:write"}, results[0].Invitation.Permissions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			permChecker := new(MockPermissionChecker)
			svc := service.NewInvitationService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo, permChecker)
			results, err := svc.ListInvitations(context.Background(), userID)
			tt.expect(t, results, err)
		})
	}
}

func TestInvitationService_AcceptInvitation(t *testing.T) {
	profileID := uuid.UUID{1}
	userID := uuid.UUID{2}
	invitationID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	otherUserID := uuid.UUID{3}

	tests := []struct {
		name   string
		setup  func(*MockProfileRepository, *MockDoseScheduleRepository, *MockPermissionChecker)
		expect func(*testing.T, *service.AcceptedProfileResult, error)
	}{
		{
			name: "AcceptInvitation_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						GrantedByUserID:  uuid.UUID{1},
						Permissions:      json.RawMessage(`["profile:read","profile:write"]`),
						Status:           "pending",
						ExpiresAt:        makeNullTime(makeExpiry(7)),
						CreatedAt:        time.Now(),
					}, nil)
				profileRepo.On("AcceptProfilePermission", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						Permissions:      json.RawMessage(`["profile:read","profile:write"]`),
						Status:           "accepted",
						UpdatedAt:        time.Now(),
					}, nil)
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile", Timezone: "UTC"}, nil)
				scheduleRepo.On("ListDoseSchedulesByProfile", mock.Anything, profileID).
					Return([]db.DoseSchedule{}, nil)
			},
			expect: func(t *testing.T, result *service.AcceptedProfileResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Test Profile", result.Profile.Name)
				assert.Equal(t, []string{"profile:read", "profile:write"}, result.Permissions)
				assert.False(t, result.Profile.IsOwner)
			},
		},
		{
			name: "AcceptInvitation_IncludeProfileAdminPermissions",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						GrantedByUserID:  uuid.UUID{1},
						Permissions:      json.RawMessage(`["profile:admin","profile:share"]`),
						Status:           "pending",
						ExpiresAt:        makeNullTime(makeExpiry(7)),
						CreatedAt:        time.Now(),
					}, nil)
				profileRepo.On("AcceptProfilePermission", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						Permissions:      json.RawMessage(`["profile:admin","profile:share"]`),
						Status:           "accepted",
					}, nil)
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test", Timezone: "UTC"}, nil)
				scheduleRepo.On("ListDoseSchedulesByProfile", mock.Anything, profileID).
					Return([]db.DoseSchedule{}, nil)
			},
			expect: func(t *testing.T, result *service.AcceptedProfileResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, []string{"profile:admin", "profile:share"}, result.Permissions)
			},
		},
		{
			name: "AcceptInvitation_NotFound",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, result *service.AcceptedProfileResult, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationNotFound)
				assert.Nil(t, result)
			},
		},
		{
			name: "AcceptInvitation_WrongUser",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: otherUserID,
						Status:           "pending",
						ExpiresAt:        makeNullTime(makeExpiry(7)),
					}, nil)
			},
			expect: func(t *testing.T, result *service.AcceptedProfileResult, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationNotFound)
				assert.Nil(t, result)
			},
		},
		{
			name: "AcceptInvitation_AlreadyAccepted",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						Status:           "accepted",
						ExpiresAt:        makeNullTime(makeExpiry(7)),
					}, nil)
			},
			expect: func(t *testing.T, result *service.AcceptedProfileResult, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationAlreadyProcessed)
				assert.Nil(t, result)
			},
		},
		{
			name: "AcceptInvitation_AlreadyDeclined",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						Status:           "declined",
						ExpiresAt:        makeNullTime(makeExpiry(7)),
					}, nil)
			},
			expect: func(t *testing.T, result *service.AcceptedProfileResult, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationAlreadyProcessed)
				assert.Nil(t, result)
			},
		},
		{
			name: "AcceptInvitation_Expired",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						Status:           "pending",
						ExpiresAt:        makeNullTime(time.Now().AddDate(0, 0, -1)),
					}, nil)
			},
			expect: func(t *testing.T, result *service.AcceptedProfileResult, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationExpired)
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			permChecker := new(MockPermissionChecker)
			svc := service.NewInvitationService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo, permChecker)
			result, err := svc.AcceptInvitation(context.Background(), invitationID, userID)
			tt.expect(t, result, err)
		})
	}
}

func TestInvitationService_DeclineInvitation(t *testing.T) {
	profileID := uuid.UUID{1}
	userID := uuid.UUID{2}
	otherUserID := uuid.UUID{3}
	invitationID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name   string
		setup  func(*MockProfileRepository, *MockDoseScheduleRepository, *MockPermissionChecker)
		expect func(*testing.T, error)
	}{
		{
			name: "DeclineInvitation_Success",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						Status:           "pending",
					}, nil)
				profileRepo.On("UpdateProfilePermissionStatus", mock.Anything, invitationID, "declined").
					Return(db.ProfilePermission{}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "DeclineInvitation_NotFound",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationNotFound)
			},
		},
		{
			name: "DeclineInvitation_WrongUser",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: otherUserID,
						Status:           "pending",
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationNotFound)
			},
		},
		{
			name: "DeclineInvitation_AlreadyAccepted",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						Status:           "accepted",
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationAlreadyProcessed)
			},
		},
		{
			name: "DeclineInvitation_AlreadyDeclined",
			setup: func(profileRepo *MockProfileRepository, scheduleRepo *MockDoseScheduleRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfilePermissionByID", mock.Anything, invitationID).
					Return(db.ProfilePermission{
						ID:               invitationID,
						ProfileID:        profileID,
						SharedWithUserID: userID,
						Status:           "declined",
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrInvitationAlreadyProcessed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			scheduleRepo := new(MockDoseScheduleRepository)
			permChecker := new(MockPermissionChecker)
			svc := service.NewInvitationService(profileRepo, scheduleRepo)
			tt.setup(profileRepo, scheduleRepo, permChecker)
			err := svc.DeclineInvitation(context.Background(), invitationID, userID)
			tt.expect(t, err)
		})
	}
}
