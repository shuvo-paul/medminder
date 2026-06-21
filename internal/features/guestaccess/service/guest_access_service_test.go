package service_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/guestaccess/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGuestAccessRepository struct {
	mock.Mock
}

func (m *MockGuestAccessRepository) Create(ctx context.Context, arg db.CreateGuestAccessTokenParams) (db.GuestAccessToken, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.GuestAccessToken), args.Error(1)
}

func (m *MockGuestAccessRepository) GetByHash(ctx context.Context, tokenHash string) (db.GuestAccessToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(db.GuestAccessToken), args.Error(1)
}

func (m *MockGuestAccessRepository) GetByID(ctx context.Context, id uuid.UUID) (db.GuestAccessToken, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.GuestAccessToken), args.Error(1)
}

func (m *MockGuestAccessRepository) ListByProfile(ctx context.Context, profileID uuid.UUID) ([]db.GuestAccessToken, error) {
	args := m.Called(ctx, profileID)
	return args.Get(0).([]db.GuestAccessToken), args.Error(1)
}

func (m *MockGuestAccessRepository) UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGuestAccessRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockPermissionChecker struct {
	mock.Mock
}

func (m *MockPermissionChecker) HasPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, profileID, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockPermissionChecker) HasAnyPermission(ctx context.Context, profileID uuid.UUID, userID uuid.UUID, permissions []string) (bool, error) {
	args := m.Called(ctx, profileID, userID, permissions)
	return args.Bool(0), args.Error(1)
}

func makeGuestAccessToken(profileID uuid.UUID, permissions []string, expiresAt time.Time) db.GuestAccessToken {
	permBytes, _ := json.Marshal(permissions)
	return db.GuestAccessToken{
		ID:          uuid.New(),
		ProfileID:   profileID,
		TokenHash:   "hash",
		Label:       sql.NullString{String: "test-token", Valid: true},
		Permissions: permBytes,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		LastUsedAt:  sql.NullTime{},
	}
}

func TestGuestAccessService_CreateToken(t *testing.T) {
	t.Run("creates_token_with_default_expiry_and_permissions", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		repo.On("Create", mock.Anything, mock.Anything).Return(
			makeGuestAccessToken(profileID, []string{"medication:read", "reminder:read"}, time.Now().AddDate(0, 0, 30)),
			nil,
		)

		result, err := svc.CreateToken(context.Background(), profileID, "", nil, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.RawToken)
		assert.Len(t, result.RawToken, 64)
		assert.WithinDuration(t, time.Now().AddDate(0, 0, 30), result.ExpiresAt, time.Second)
		repo.AssertExpectations(t)
	})

	t.Run("creates_token_with_custom_permissions_and_expiry", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		repo.On("Create", mock.Anything, mock.Anything).Return(
			makeGuestAccessToken(profileID, []string{"medication:read"}, time.Now().AddDate(0, 0, 7)),
			nil,
		)

		result, err := svc.CreateToken(context.Background(), profileID, "my-token", []string{"medication:read"}, 7)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "my-token", result.Label)
		assert.WithinDuration(t, time.Now().AddDate(0, 0, 7), result.ExpiresAt, time.Second)
		repo.AssertExpectations(t)
	})

	t.Run("returns_error_on_repo_failure", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		repo.On("Create", mock.Anything, mock.Anything).Return(db.GuestAccessToken{}, assert.AnError)

		_, err := svc.CreateToken(context.Background(), profileID, "", nil, 30)

		assert.Error(t, err)
	})
}

func TestGuestAccessService_ListTokens(t *testing.T) {
	t.Run("returns_tokens_for_profile", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		token := makeGuestAccessToken(profileID, []string{"medication:read"}, time.Now().AddDate(0, 0, 30))
		repo.On("ListByProfile", mock.Anything, profileID).Return([]db.GuestAccessToken{token}, nil)

		results, err := svc.ListTokens(context.Background(), profileID)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, token.ID, results[0].ID)
	})

	t.Run("returns_empty_list_when_no_tokens", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		repo.On("ListByProfile", mock.Anything, profileID).Return([]db.GuestAccessToken{}, nil)

		results, err := svc.ListTokens(context.Background(), profileID)

		assert.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestGuestAccessService_RevokeToken(t *testing.T) {
	t.Run("revokes_token_when_authorized", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		userID := uuid.New()
		tokenID := uuid.New()
		token := makeGuestAccessToken(profileID, []string{"medication:read"}, time.Now().AddDate(0, 0, 30))
		token.ID = tokenID

		repo.On("GetByID", mock.Anything, tokenID).Return(token, nil)
		permChecker.On("HasAnyPermission", mock.Anything, profileID, userID, []string{"profile:admin", "profile:owner"}).Return(true, nil)
		repo.On("Delete", mock.Anything, tokenID).Return(nil)

		err := svc.RevokeToken(context.Background(), tokenID, userID)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		permChecker.AssertExpectations(t)
	})

	t.Run("returns_error_when_not_authorized", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		userID := uuid.New()
		tokenID := uuid.New()
		token := makeGuestAccessToken(profileID, []string{"medication:read"}, time.Now().AddDate(0, 0, 30))
		token.ID = tokenID

		repo.On("GetByID", mock.Anything, tokenID).Return(token, nil)
		permChecker.On("HasAnyPermission", mock.Anything, profileID, userID, []string{"profile:admin", "profile:owner"}).Return(false, nil)

		err := svc.RevokeToken(context.Background(), tokenID, userID)

		assert.ErrorIs(t, err, service.ErrGuestTokenInsufficientPerms)
	})

	t.Run("returns_error_when_token_not_found", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		tokenID := uuid.New()
		repo.On("GetByID", mock.Anything, tokenID).Return(db.GuestAccessToken{}, sql.ErrNoRows)

		err := svc.RevokeToken(context.Background(), tokenID, uuid.New())

		assert.ErrorIs(t, err, service.ErrGuestTokenNotFound)
	})
}

func TestGuestAccessService_Authenticate(t *testing.T) {
	t.Run("authenticates_valid_token", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		rawToken := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		token := makeGuestAccessToken(profileID, []string{"medication:read", "reminder:read"}, time.Now().AddDate(0, 0, 30))

		repo.On("GetByHash", mock.Anything, mock.Anything).Return(token, nil)
		repo.On("UpdateLastUsedAt", mock.Anything, token.ID).Return(nil)

		result, err := svc.Authenticate(context.Background(), rawToken)

		assert.NoError(t, err)
		assert.Equal(t, profileID, result.ProfileID)
		assert.Equal(t, token.ID, result.TokenID)
		assert.Contains(t, result.Permissions, "medication:read")
		assert.Contains(t, result.Permissions, "reminder:read")
	})

	t.Run("returns_error_for_expired_token", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		profileID := uuid.New()
		expiredToken := makeGuestAccessToken(profileID, []string{"medication:read"}, time.Now().Add(-time.Hour))

		repo.On("GetByHash", mock.Anything, mock.Anything).Return(expiredToken, nil)

		_, err := svc.Authenticate(context.Background(), "some-token")

		assert.ErrorIs(t, err, service.ErrGuestTokenExpired)
	})

	t.Run("returns_error_for_nonexistent_token", func(t *testing.T) {
		repo := new(MockGuestAccessRepository)
		permChecker := new(MockPermissionChecker)
		svc := service.NewGuestAccessService(repo, permChecker)

		repo.On("GetByHash", mock.Anything, mock.Anything).Return(db.GuestAccessToken{}, sql.ErrNoRows)

		_, err := svc.Authenticate(context.Background(), "nonexistent-token")

		assert.ErrorIs(t, err, service.ErrGuestTokenNotFound)
	})
}
