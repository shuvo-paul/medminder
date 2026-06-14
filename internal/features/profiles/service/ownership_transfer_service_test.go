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

type MockOwnershipTransferRepository struct {
	mock.Mock
}

func (m *MockOwnershipTransferRepository) CreateTransfer(ctx context.Context, profileID, fromUserID, toUserID uuid.UUID, expiresAt time.Time) (db.OwnershipTransfer, error) {
	args := m.Called(ctx, profileID, fromUserID, toUserID, expiresAt)
	return args.Get(0).(db.OwnershipTransfer), args.Error(1)
}

func (m *MockOwnershipTransferRepository) GetTransferByID(ctx context.Context, transferID uuid.UUID) (db.OwnershipTransfer, error) {
	args := m.Called(ctx, transferID)
	return args.Get(0).(db.OwnershipTransfer), args.Error(1)
}

func (m *MockOwnershipTransferRepository) GetPendingTransferByProfile(ctx context.Context, profileID uuid.UUID) (db.OwnershipTransfer, error) {
	args := m.Called(ctx, profileID)
	return args.Get(0).(db.OwnershipTransfer), args.Error(1)
}

func (m *MockOwnershipTransferRepository) ListPendingTransfersByUser(ctx context.Context, userID uuid.UUID) ([]db.OwnershipTransfer, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.OwnershipTransfer), args.Error(1)
}

func (m *MockOwnershipTransferRepository) UpdateTransferStatus(ctx context.Context, transferID uuid.UUID, status string) (db.OwnershipTransfer, error) {
	args := m.Called(ctx, transferID, status)
	return args.Get(0).(db.OwnershipTransfer), args.Error(1)
}

func (m *MockOwnershipTransferRepository) ListPendingTransfersWithDetailsByUser(ctx context.Context, userID uuid.UUID) ([]db.ListPendingTransfersWithDetailsByUserRow, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.ListPendingTransfersWithDetailsByUserRow), args.Error(1)
}

func TestInitiateTransfer(t *testing.T) {
	profileID := uuid.UUID{1}
	fromUserID := uuid.UUID{2}
	toUserID := uuid.UUID{3}
	now := time.Now()

	tests := []struct {
		name             string
		setup            func(*MockProfileRepository, *MockOwnershipTransferRepository, *MockPermissionChecker)
		expect           func(*testing.T, *service.OwnershipTransferResult, error)
		overrideFromUser uuid.UUID
		overrideToUser   uuid.UUID
	}{
		{
			name: "InitiateTransfer_Success",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile"}, nil)
				profileRepo.On("UserExists", mock.Anything, toUserID).Return(true, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, profileID, toUserID).
					Return(db.ProfilePermission{
						ProfileID:   profileID,
						Permissions: json.RawMessage(`["profile:admin","profile:read"]`),
						Status:      "accepted",
					}, nil)
				transferRepo.On("GetPendingTransferByProfile", mock.Anything, profileID).
					Return(db.OwnershipTransfer{}, sql.ErrNoRows)
				transferRepo.On("CreateTransfer", mock.Anything, profileID, fromUserID, toUserID, mock.AnythingOfType("time.Time")).
					Return(db.OwnershipTransfer{
						ID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, 7),
						CreatedAt:  now,
					}, nil)
				profileRepo.On("GetUserByID", mock.Anything, fromUserID).
					Return(db.User{ID: fromUserID, DisplayName: "From User"}, nil)
				profileRepo.On("GetUserByID", mock.Anything, toUserID).
					Return(db.User{ID: toUserID, DisplayName: "To User"}, nil)
			},
			expect: func(t *testing.T, result *service.OwnershipTransferResult, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "pending", result.Transfer.Status)
				assert.Equal(t, profileID, result.Transfer.ProfileID)
				assert.Equal(t, "From User", result.Transfer.FromName)
				assert.Equal(t, "To User", result.Transfer.ToName)
			},
		},
		{
			name: "InitiateTransfer_ProfileNotFound",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, result *service.OwnershipTransferResult, err error) {
				assert.ErrorIs(t, err, service.ErrProfileNotFound)
				assert.Nil(t, result)
			},
		},
		{
			name: "InitiateTransfer_SelfTransfer",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile"}, nil)
			},
			expect: func(t *testing.T, result *service.OwnershipTransferResult, err error) {
				assert.ErrorIs(t, err, service.ErrCannotTransferToSelf)
				assert.Nil(t, result)
			},
			overrideFromUser: fromUserID,
			overrideToUser:   fromUserID,
		},
		{
			name: "InitiateTransfer_UserNotFound",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile"}, nil)
				profileRepo.On("UserExists", mock.Anything, toUserID).Return(false, nil)
			},
			expect: func(t *testing.T, result *service.OwnershipTransferResult, err error) {
				assert.ErrorIs(t, err, service.ErrUserNotFound)
				assert.Nil(t, result)
			},
		},
		{
			name: "InitiateTransfer_NewOwnerNoAdminPermission",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile"}, nil)
				profileRepo.On("UserExists", mock.Anything, toUserID).Return(true, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, profileID, toUserID).
					Return(db.ProfilePermission{
						ProfileID:   profileID,
						Permissions: json.RawMessage(`["profile:read"]`),
						Status:      "accepted",
					}, nil)
			},
			expect: func(t *testing.T, result *service.OwnershipTransferResult, err error) {
				assert.ErrorIs(t, err, service.ErrNewOwnerNoAdminPermission)
				assert.Nil(t, result)
			},
		},
		{
			name: "InitiateTransfer_PendingTransferExists",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile"}, nil)
				profileRepo.On("UserExists", mock.Anything, toUserID).Return(true, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, profileID, toUserID).
					Return(db.ProfilePermission{
						ProfileID:   profileID,
						Permissions: json.RawMessage(`["profile:admin"]`),
						Status:      "accepted",
					}, nil)
				transferRepo.On("GetPendingTransferByProfile", mock.Anything, profileID).
					Return(db.OwnershipTransfer{
						ID:        uuid.MustParse("00000000-0000-0000-0000-000000000002"),
						ProfileID: profileID,
						Status:    "pending",
					}, nil)
			},
			expect: func(t *testing.T, result *service.OwnershipTransferResult, err error) {
				assert.ErrorIs(t, err, service.ErrPendingTransferExists)
				assert.Nil(t, result)
			},
		},
		{
			name: "InitiateTransfer_NewOwnerPermissionNotAccepted",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				profileRepo.On("GetProfileByID", mock.Anything, profileID).
					Return(db.Profile{ID: profileID, Name: "Test Profile"}, nil)
				profileRepo.On("UserExists", mock.Anything, toUserID).Return(true, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, profileID, toUserID).
					Return(db.ProfilePermission{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, result *service.OwnershipTransferResult, err error) {
				assert.ErrorIs(t, err, service.ErrNewOwnerNoAdminPermission)
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			transferRepo := new(MockOwnershipTransferRepository)
			svc := service.NewOwnershipTransferService(profileRepo, transferRepo)
			tt.setup(profileRepo, transferRepo, nil)

			fu := fromUserID
			tu := toUserID
			if tt.overrideFromUser != uuid.Nil {
				fu = tt.overrideFromUser
			}
			if tt.overrideToUser != uuid.Nil {
				tu = tt.overrideToUser
			}
			result, err := svc.InitiateTransfer(context.Background(), profileID, fu, tu)
			tt.expect(t, result, err)
		})
	}
}

func TestListPendingTransfers(t *testing.T) {
	userID := uuid.UUID{1}
	profileID := uuid.UUID{2}
	fromUserID := uuid.UUID{3}
	now := time.Now()

	tests := []struct {
		name   string
		setup  func(*MockProfileRepository, *MockOwnershipTransferRepository, *MockPermissionChecker)
		expect func(*testing.T, []service.OwnershipTransferResult, error)
	}{
		{
			name: "ListPendingTransfers_Success",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("ListPendingTransfersWithDetailsByUser", mock.Anything, userID).
					Return([]db.ListPendingTransfersWithDetailsByUserRow{
						{
							ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
							ProfileID:   profileID,
							FromUserID:  fromUserID,
							ToUserID:    userID,
							Status:      "pending",
							ExpiresAt:   now.AddDate(0, 0, 7),
							CreatedAt:   now,
							ProfileName: "Test Profile",
							FromName:    "From User",
							ToName:      "To User",
						},
					}, nil)
			},
			expect: func(t *testing.T, results []service.OwnershipTransferResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, "Test Profile", results[0].Transfer.ProfileName)
				assert.Equal(t, "From User", results[0].Transfer.FromName)
				assert.Equal(t, "To User", results[0].Transfer.ToName)
				assert.Equal(t, "pending", results[0].Transfer.Status)
			},
		},
		{
			name: "ListPendingTransfers_Empty",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("ListPendingTransfersWithDetailsByUser", mock.Anything, userID).
					Return([]db.ListPendingTransfersWithDetailsByUserRow{}, nil)
			},
			expect: func(t *testing.T, results []service.OwnershipTransferResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			transferRepo := new(MockOwnershipTransferRepository)
			svc := service.NewOwnershipTransferService(profileRepo, transferRepo)
			tt.setup(profileRepo, transferRepo, nil)
			results, err := svc.ListPendingTransfers(context.Background(), userID)
			tt.expect(t, results, err)
		})
	}
}

func TestAcceptTransfer(t *testing.T) {
	transferID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	profileID := uuid.UUID{1}
	fromUserID := uuid.UUID{2}
	toUserID := uuid.UUID{3}
	now := time.Now()

	tests := []struct {
		name   string
		setup  func(*MockProfileRepository, *MockOwnershipTransferRepository, *MockPermissionChecker)
		expect func(*testing.T, error)
	}{
		{
			name: "AcceptTransfer_Success",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, profileID, toUserID).
					Return(db.ProfilePermission{
						ID:               uuid.MustParse("00000000-0000-0000-0000-000000000010"),
						ProfileID:        profileID,
						SharedWithUserID: toUserID,
						Permissions:      json.RawMessage(`["profile:admin"]`),
						Status:           "accepted",
					}, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, profileID, fromUserID).
					Return(db.ProfilePermission{
						ID:               uuid.MustParse("00000000-0000-0000-0000-000000000011"),
						ProfileID:        profileID,
						SharedWithUserID: fromUserID,
						Permissions:      json.RawMessage(`["profile:owner","profile:admin"]`),
						Status:           "accepted",
					}, nil)
				profileRepo.On("UpdateProfilePermissionByProfileAndUser", mock.Anything, profileID, fromUserID, mock.AnythingOfType("json.RawMessage")).
					Return(db.ProfilePermission{}, nil)
				profileRepo.On("UpdateProfilePermissionByProfileAndUser", mock.Anything, profileID, toUserID, mock.AnythingOfType("json.RawMessage")).
					Return(db.ProfilePermission{}, nil)
				transferRepo.On("UpdateTransferStatus", mock.Anything, transferID, "accepted").
					Return(db.OwnershipTransfer{}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "AcceptTransfer_NotFound",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotFound)
			},
		},
		{
			name: "AcceptTransfer_NotRecipient",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   uuid.UUID{9},
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotFound)
			},
		},
		{
			name: "AcceptTransfer_AlreadyAccepted",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "accepted",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotPending)
			},
		},
		{
			name: "AcceptTransfer_AlreadyDeclined",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "declined",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotPending)
			},
		},
		{
			name: "AcceptTransfer_Expired",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, -1),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferExpired)
			},
		},
		{
			name: "AcceptTransfer_NewOwnerNoLongerHasAdmin",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
				profileRepo.On("GetProfilePermission", mock.Anything, profileID, toUserID).
					Return(db.ProfilePermission{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrNewOwnerNoAdminPermission)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			transferRepo := new(MockOwnershipTransferRepository)
			svc := service.NewOwnershipTransferService(profileRepo, transferRepo)
			tt.setup(profileRepo, transferRepo, nil)
			err := svc.AcceptTransfer(context.Background(), transferID, toUserID)
			tt.expect(t, err)
		})
	}
}

func TestDeclineTransfer(t *testing.T) {
	transferID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	profileID := uuid.UUID{1}
	fromUserID := uuid.UUID{2}
	toUserID := uuid.UUID{3}
	otherUserID := uuid.UUID{9}
	now := time.Now()

	tests := []struct {
		name   string
		setup  func(*MockProfileRepository, *MockOwnershipTransferRepository, *MockPermissionChecker)
		expect func(*testing.T, error)
	}{
		{
			name: "DeclineTransfer_Success",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
				transferRepo.On("UpdateTransferStatus", mock.Anything, transferID, "declined").
					Return(db.OwnershipTransfer{
						ID:     transferID,
						Status: "declined",
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "DeclineTransfer_NotFound",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotFound)
			},
		},
		{
			name: "DeclineTransfer_WrongUser",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   otherUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotFound)
			},
		},
		{
			name: "DeclineTransfer_AlreadyAccepted",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "accepted",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotPending)
			},
		},
		{
			name: "DeclineTransfer_Expired",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, -1),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferExpired)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			transferRepo := new(MockOwnershipTransferRepository)
			svc := service.NewOwnershipTransferService(profileRepo, transferRepo)
			tt.setup(profileRepo, transferRepo, nil)
			err := svc.DeclineTransfer(context.Background(), transferID, toUserID)
			tt.expect(t, err)
		})
	}
}

func TestCancelTransfer(t *testing.T) {
	transferID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	profileID := uuid.UUID{1}
	fromUserID := uuid.UUID{2}
	toUserID := uuid.UUID{3}
	otherUserID := uuid.UUID{9}
	now := time.Now()

	tests := []struct {
		name         string
		setup        func(*MockProfileRepository, *MockOwnershipTransferRepository, *MockPermissionChecker)
		expect       func(*testing.T, error)
		actingUserID uuid.UUID
	}{
		{
			name: "CancelTransfer_Success",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
				transferRepo.On("UpdateTransferStatus", mock.Anything, transferID, "cancelled").
					Return(db.OwnershipTransfer{
						ID:     transferID,
						Status: "cancelled",
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			actingUserID: fromUserID,
		},
		{
			name: "CancelTransfer_NotFound",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{}, sql.ErrNoRows)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotFound)
			},
			actingUserID: fromUserID,
		},
		{
			name: "CancelTransfer_WrongUser",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotInitiator)
			},
			actingUserID: otherUserID,
		},
		{
			name: "CancelTransfer_AlreadyAccepted",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "accepted",
						ExpiresAt:  now.AddDate(0, 0, 7),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferNotPending)
			},
			actingUserID: fromUserID,
		},
		{
			name: "CancelTransfer_Expired",
			setup: func(profileRepo *MockProfileRepository, transferRepo *MockOwnershipTransferRepository, permChecker *MockPermissionChecker) {
				transferRepo.On("GetTransferByID", mock.Anything, transferID).
					Return(db.OwnershipTransfer{
						ID:         transferID,
						ProfileID:  profileID,
						FromUserID: fromUserID,
						ToUserID:   toUserID,
						Status:     "pending",
						ExpiresAt:  now.AddDate(0, 0, -1),
					}, nil)
			},
			expect: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, service.ErrTransferExpired)
			},
			actingUserID: fromUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := new(MockProfileRepository)
			transferRepo := new(MockOwnershipTransferRepository)
			svc := service.NewOwnershipTransferService(profileRepo, transferRepo)
			tt.setup(profileRepo, transferRepo, nil)

			err := svc.CancelTransfer(context.Background(), transferID, tt.actingUserID)
			tt.expect(t, err)
		})
	}
}
