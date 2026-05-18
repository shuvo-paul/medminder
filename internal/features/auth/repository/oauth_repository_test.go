package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/common/database/testutil"
	sqlc "github.com/shuvo-paul/medminder/internal/database/sqlc"
	"github.com/shuvo-paul/medminder/internal/features/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests using testcontainers — run with the testutil package.
// These tests verify the actual repository behavior against a real PostgreSQL database.

func TestOAuthAccountRepository_Integration(t *testing.T) {
	tc := testutil.SetupPostgresContainer(t)
	defer tc.Teardown(t)

	dbConn := tc.Connect(t)
	defer dbConn.Close()

	m := tc.NewMigrator(t)
	err := m.Up()
	require.NoError(t, err, "should run migrations up")

	queries := sqlc.New(dbConn)
	repo := repository.NewOAuthAccountRepository(queries)

	// Helper to create a test user for OAuth account tests
	createTestUser := func(t *testing.T) uuid.UUID {
		t.Helper()
		email := uuid.New().String() + "@test.com"
		row, err := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
			Email:         email,
			DisplayName:   "Test User",
			PasswordHash:  sql.NullString{String: "hash", Valid: true},
			EmailVerified: sql.NullBool{Bool: false, Valid: true},
		})
		require.NoError(t, err, "should create test user")
		return row.ID
	}

	t.Run("CreateOAuthAccount_Success", func(t *testing.T) {
		id := uuid.New()
		userID := createTestUser(t)

		result, err := repo.CreateOAuthAccount(context.Background(), id, userID, "google", "google123")

		assert.NoError(t, err)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, "google", result.Provider)
		assert.Equal(t, "google123", result.ProviderUserID)
	})

	t.Run("CreateOAuthAccount_Duplicate_UniqueConstraint", func(t *testing.T) {
		userID := createTestUser(t)

		// Create first account
		_, err := repo.CreateOAuthAccount(context.Background(), uuid.New(), userID, "github", "github123")
		require.NoError(t, err)

		// Try to create duplicate (same provider for same user)
		_, err = repo.CreateOAuthAccount(context.Background(), uuid.New(), userID, "github", "github456")

		assert.Error(t, err, "should fail with unique constraint violation")
	})

	t.Run("GetOAuthAccountByProviderAndUserID_Success", func(t *testing.T) {
		id := uuid.New()
		userID := createTestUser(t)
		provider := "google"
		providerUserID := "get-by-provider-user-test"

		// Create account
		_, err := repo.CreateOAuthAccount(context.Background(), id, userID, provider, providerUserID)
		require.NoError(t, err)

		// Retrieve by provider + provider user ID
		result, err := repo.GetOAuthAccountByProviderAndUserID(context.Background(), provider, providerUserID)

		assert.NoError(t, err)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, provider, result.Provider)
	})

	t.Run("GetOAuthAccountByProviderAndUserID_NotFound", func(t *testing.T) {
		_, err := repo.GetOAuthAccountByProviderAndUserID(context.Background(), "unknown-provider", "unknown-user")

		assert.Error(t, err)
		assert.Equal(t, repository.ErrOAuthAccountNotFound, err)
	})

	t.Run("GetOAuthAccountByUserIDAndProvider_Success", func(t *testing.T) {
		id := uuid.New()
		userID := createTestUser(t)
		provider := "linkedin"
		providerUserID := "linkedin-user-123"

		// Create account
		_, err := repo.CreateOAuthAccount(context.Background(), id, userID, provider, providerUserID)
		require.NoError(t, err)

		// Retrieve by user ID + provider
		result, err := repo.GetOAuthAccountByUserIDAndProvider(context.Background(), userID, provider)

		assert.NoError(t, err)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, provider, result.Provider)
		assert.Equal(t, providerUserID, result.ProviderUserID)
	})

	t.Run("GetOAuthAccountByUserIDAndProvider_NotFound", func(t *testing.T) {
		_, err := repo.GetOAuthAccountByUserIDAndProvider(context.Background(), uuid.New(), "unknown-provider")

		assert.Error(t, err)
		assert.Equal(t, repository.ErrOAuthAccountNotFound, err)
	})

	t.Run("GetOAuthAccountsByUserID_Success_Multiple", func(t *testing.T) {
		userID := createTestUser(t)

		// Create multiple linked accounts
		_, err := repo.CreateOAuthAccount(context.Background(), uuid.New(), userID, "google", "multi-google")
		require.NoError(t, err)
		_, err = repo.CreateOAuthAccount(context.Background(), uuid.New(), userID, "github", "multi-github")
		require.NoError(t, err)

		// Retrieve all accounts for user
		accounts, err := repo.GetOAuthAccountsByUserID(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, accounts, 2)
	})

	t.Run("GetOAuthAccountsByUserID_Empty", func(t *testing.T) {
		accounts, err := repo.GetOAuthAccountsByUserID(context.Background(), uuid.New())

		assert.NoError(t, err)
		assert.Empty(t, accounts)
	})

	t.Run("DeleteOAuthAccount_Success", func(t *testing.T) {
		id := uuid.New()
		userID := createTestUser(t)

		// Create account
		_, err := repo.CreateOAuthAccount(context.Background(), id, userID, "delete-test", "delete-user")
		require.NoError(t, err)

		// Delete account
		deleted, err := repo.DeleteOAuthAccount(context.Background(), id)

		assert.NoError(t, err)
		assert.Equal(t, id, deleted.ID)

		// Verify deleted
		_, err = repo.GetOAuthAccountByProviderAndUserID(context.Background(), "delete-test", "delete-user")
		assert.Equal(t, repository.ErrOAuthAccountNotFound, err)
	})

	t.Run("DeleteOAuthAccount_NotFound", func(t *testing.T) {
		_, err := repo.DeleteOAuthAccount(context.Background(), uuid.New())

		assert.Error(t, err)
	})

	t.Run("DeleteOAuthAccountByUserIDAndProvider_Success", func(t *testing.T) {
		userID := createTestUser(t)

		// Create account
		_, err := repo.CreateOAuthAccount(context.Background(), uuid.New(), userID, "unlink-test", "unlink-user")
		require.NoError(t, err)

		// Unlink provider
		err = repo.DeleteOAuthAccountByUserIDAndProvider(context.Background(), userID, "unlink-test")

		assert.NoError(t, err)

		// Verify unlinked
		_, err = repo.GetOAuthAccountByUserIDAndProvider(context.Background(), userID, "unlink-test")
		assert.Equal(t, repository.ErrOAuthAccountNotFound, err)
	})
}

func TestOAuthAuthorizationCodeRepository_Integration(t *testing.T) {
	tc := testutil.SetupPostgresContainer(t)
	defer tc.Teardown(t)

	dbConn := tc.Connect(t)
	defer dbConn.Close()

	m := tc.NewMigrator(t)
	err := m.Up()
	require.NoError(t, err, "should run migrations up")

	queries := sqlc.New(dbConn)
	repo := repository.NewOAuthAuthorizationCodeRepository(dbConn, queries)

	// Helper to create a test user for OAuth authorization code tests
	createTestUser := func(t *testing.T) uuid.UUID {
		t.Helper()
		email := uuid.New().String() + "@test.com"
		row, err := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
			Email:         email,
			DisplayName:   "Test User",
			PasswordHash:  sql.NullString{String: "hash", Valid: true},
			EmailVerified: sql.NullBool{Bool: false, Valid: true},
		})
		require.NoError(t, err, "should create test user")
		return row.ID
	}

	t.Run("CreateAuthorizationCode_Success", func(t *testing.T) {
		id := uuid.New()
		codeHash := "test-code-hash-123"
		userID := uuid.NullUUID{UUID: createTestUser(t), Valid: true}
		nonce := "test-nonce"
		purpose := "login"
		expiresAt := time.Now().Add(10 * time.Minute)

		result, err := repo.CreateAuthorizationCode(context.Background(), id, codeHash, userID, nonce, purpose, expiresAt)

		assert.NoError(t, err)
		assert.Equal(t, codeHash, result.CodeHash)
		assert.Equal(t, nonce, result.Nonce)
		assert.Equal(t, purpose, result.Purpose)
		assert.Equal(t, userID.UUID, result.UserID.UUID)
	})

	t.Run("GetAuthorizationCodeByHash_Success", func(t *testing.T) {
		id := uuid.New()
		codeHash := "get-by-hash-test"
		userID := uuid.NullUUID{UUID: createTestUser(t), Valid: true}
		nonce := "nonce123"
		purpose := "register"
		expiresAt := time.Now().Add(15 * time.Minute)

		// Create code
		_, err := repo.CreateAuthorizationCode(context.Background(), id, codeHash, userID, nonce, purpose, expiresAt)
		require.NoError(t, err)

		// Retrieve by hash
		result, err := repo.GetAuthorizationCodeByHash(context.Background(), codeHash)

		assert.NoError(t, err)
		assert.Equal(t, codeHash, result.CodeHash)
		assert.Equal(t, nonce, result.Nonce)
		assert.Equal(t, purpose, result.Purpose)
		assert.False(t, result.UsedAt.Valid, "should not be marked as used yet")
	})

	t.Run("GetAuthorizationCodeByHash_NotFound", func(t *testing.T) {

		_, err := repo.GetAuthorizationCodeByHash(context.Background(), "non-existent-hash")

		assert.Error(t, err)
		assert.Equal(t, repository.ErrOAuthCodeNotFound, err)
	})

	t.Run("GetAndLockAuthorizationCode_Success", func(t *testing.T) {
		id := uuid.New()
		codeHash := "lock-test-success"
		userID := uuid.NullUUID{UUID: createTestUser(t), Valid: true}
		nonce := "lock-nonce"
		purpose := "login"
		expiresAt := time.Now().Add(10 * time.Minute)

		// Create code
		_, err := repo.CreateAuthorizationCode(context.Background(), id, codeHash, userID, nonce, purpose, expiresAt)
		require.NoError(t, err)

		// Lock and retrieve
		result, err := repo.GetAndLockAuthorizationCode(context.Background(), codeHash)

		assert.NoError(t, err)
		assert.Equal(t, codeHash, result.CodeHash)
	})

	t.Run("GetAndLockAuthorizationCode_NotFound", func(t *testing.T) {

		_, err := repo.GetAndLockAuthorizationCode(context.Background(), "non-existent-lock-hash")

		assert.Error(t, err)
		assert.Equal(t, repository.ErrOAuthCodeNotFound, err)
	})

	t.Run("GetAndLockAuthorizationCode_AlreadyUsed_ReturnsNotFound", func(t *testing.T) {
		id := uuid.New()
		codeHash := "lock-test-used"
		userID := uuid.NullUUID{UUID: createTestUser(t), Valid: true}
		nonce := "used-nonce"
		purpose := "link"
		expiresAt := time.Now().Add(10 * time.Minute)

		// Create code
		_, err := repo.CreateAuthorizationCode(context.Background(), id, codeHash, userID, nonce, purpose, expiresAt)
		require.NoError(t, err)

		// Mark as used
		_, err = repo.MarkAuthorizationCodeAsUsed(context.Background(), codeHash)
		require.NoError(t, err)

		// Try to lock (should fail with same error as not found)
		_, err = repo.GetAndLockAuthorizationCode(context.Background(), codeHash)

		assert.Error(t, err)
		assert.Equal(t, repository.ErrOAuthCodeNotFound, err)
	})

	t.Run("GetAndLockAuthorizationCode_Expired_ReturnsNotFound", func(t *testing.T) {
		id := uuid.New()
		codeHash := "lock-test-expired"
		userID := uuid.NullUUID{UUID: createTestUser(t), Valid: true}
		nonce := "expired-nonce"
		purpose := "login"
		expiresAt := time.Now().Add(-1 * time.Hour) // Already expired

		// Create expired code
		_, err := repo.CreateAuthorizationCode(context.Background(), id, codeHash, userID, nonce, purpose, expiresAt)
		require.NoError(t, err)

		// Try to lock (should fail with same error as not found)
		_, err = repo.GetAndLockAuthorizationCode(context.Background(), codeHash)

		assert.Error(t, err)
		assert.Equal(t, repository.ErrOAuthCodeNotFound, err)
	})

	t.Run("MarkAuthorizationCodeAsUsed_Success", func(t *testing.T) {
		id := uuid.New()
		codeHash := "mark-used-test"
		userID := uuid.NullUUID{UUID: createTestUser(t), Valid: true}
		nonce := "mark-nonce"
		purpose := "login"
		expiresAt := time.Now().Add(10 * time.Minute)

		// Create code
		_, err := repo.CreateAuthorizationCode(context.Background(), id, codeHash, userID, nonce, purpose, expiresAt)
		require.NoError(t, err)

		// Mark as used
		result, err := repo.MarkAuthorizationCodeAsUsed(context.Background(), codeHash)

		assert.NoError(t, err)
		assert.True(t, result.UsedAt.Valid, "should be marked as used")
	})

	t.Run("MarkAuthorizationCodeAsUsed_AlreadyUsed", func(t *testing.T) {
		id := uuid.New()
		codeHash := "mark-used-twice"
		userID := uuid.NullUUID{UUID: createTestUser(t), Valid: true}
		nonce := "twice-nonce"
		purpose := "login"
		expiresAt := time.Now().Add(10 * time.Minute)

		// Create code
		_, err := repo.CreateAuthorizationCode(context.Background(), id, codeHash, userID, nonce, purpose, expiresAt)
		require.NoError(t, err)

		// Mark as used first time
		_, err = repo.MarkAuthorizationCodeAsUsed(context.Background(), codeHash)
		require.NoError(t, err)

		// Try to mark again (should fail - no row returned because used_at is not null)
		_, err = repo.MarkAuthorizationCodeAsUsed(context.Background(), codeHash)

		assert.Error(t, err)
		assert.Equal(t, repository.ErrOAuthCodeNotFound, err)
	})

	t.Run("MarkAuthorizationCodeAsUsed_NotFound", func(t *testing.T) {

		_, err := repo.MarkAuthorizationCodeAsUsed(context.Background(), "non-existent-mark")

		assert.Error(t, err)
		assert.Equal(t, repository.ErrOAuthCodeNotFound, err)
	})

	t.Run("CleanupExpiredAuthorizationCodes_Success", func(t *testing.T) {
		userID := uuid.NullUUID{UUID: createTestUser(t), Valid: true}
		nonce := "cleanup-nonce"
		purpose := "login"

		// Create expired codes
		expiredTime := time.Now().Add(-1 * time.Hour)
		_, err := repo.CreateAuthorizationCode(context.Background(), uuid.New(), "expired-1", userID, nonce, purpose, expiredTime)
		require.NoError(t, err)
		_, err = repo.CreateAuthorizationCode(context.Background(), uuid.New(), "expired-2", userID, nonce, purpose, expiredTime)
		require.NoError(t, err)

		// Create valid code
		validTime := time.Now().Add(1 * time.Hour)
		_, err = repo.CreateAuthorizationCode(context.Background(), uuid.New(), "valid-code", userID, nonce, purpose, validTime)
		require.NoError(t, err)

		// Cleanup
		err = repo.CleanupExpiredAuthorizationCodes(context.Background())

		assert.NoError(t, err)

		// Verify expired codes are gone
		_, err = repo.GetAuthorizationCodeByHash(context.Background(), "expired-1")
		assert.Equal(t, repository.ErrOAuthCodeNotFound, err)

		// Verify valid code still exists
		_, err = repo.GetAuthorizationCodeByHash(context.Background(), "valid-code")
		assert.NoError(t, err)
	})

	t.Run("CleanupExpiredAuthorizationCodes_Empty", func(t *testing.T) {

		// Cleanup with no expired codes
		err := repo.CleanupExpiredAuthorizationCodes(context.Background())

		assert.NoError(t, err, "should not error even when nothing to clean")
	})
}

func TestOAuthErrors_Integration(t *testing.T) {
	t.Run("ErrOAuthAccountNotFound", func(t *testing.T) {
		assert.Error(t, repository.ErrOAuthAccountNotFound)
		assert.Equal(t, "oauth account not found", repository.ErrOAuthAccountNotFound.Error())
	})

	t.Run("ErrOAuthCodeNotFound", func(t *testing.T) {
		assert.Error(t, repository.ErrOAuthCodeNotFound)
		assert.Equal(t, "oauth code not found", repository.ErrOAuthCodeNotFound.Error())
	})

	t.Run("ErrorsAreDistinct", func(t *testing.T) {
		assert.NotEqual(t, repository.ErrOAuthAccountNotFound, repository.ErrOAuthCodeNotFound)
		assert.False(t, errors.Is(repository.ErrOAuthCodeNotFound, repository.ErrOAuthAccountNotFound))
	})
}
