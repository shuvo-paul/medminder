package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
)

// OAuthAccountRepository defines the interface for OAuth account data access.
type OAuthAccountRepository interface {
	CreateOAuthAccount(ctx context.Context, id uuid.UUID, userID uuid.UUID, provider string, providerUserID string) (db.OauthAccount, error)
	GetOAuthAccountByProviderAndUserID(ctx context.Context, provider string, providerUserID string) (db.OauthAccount, error)
	GetOAuthAccountByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) (db.OauthAccount, error)
	GetOAuthAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]db.OauthAccount, error)
	DeleteOAuthAccount(ctx context.Context, id uuid.UUID) (db.OauthAccount, error)
	DeleteOAuthAccountByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) error
}

// OAuthAuthorizationCodeRepository defines the interface for OAuth authorization code data access.
type OAuthAuthorizationCodeRepository interface {
	CreateAuthorizationCode(ctx context.Context, id uuid.UUID, codeHash string, userID uuid.NullUUID, nonce string, purpose string, expiresAt time.Time) (db.OauthAuthorizationCode, error)
	GetAuthorizationCodeByHash(ctx context.Context, codeHash string) (db.OauthAuthorizationCode, error)
	GetAndLockAuthorizationCode(ctx context.Context, codeHash string) (db.OauthAuthorizationCode, error)
	MarkAuthorizationCodeAsUsed(ctx context.Context, codeHash string) (db.OauthAuthorizationCode, error)
	CleanupExpiredAuthorizationCodes(ctx context.Context) error
}

// oauthAccountRepository implements OAuthAccountRepository.
type oauthAccountRepository struct {
	queries *db.Queries
}

// NewOAuthAccountRepository creates a new OAuth account repository.
func NewOAuthAccountRepository(queries *db.Queries) OAuthAccountRepository {
	return &oauthAccountRepository{queries: queries}
}

func (r *oauthAccountRepository) CreateOAuthAccount(ctx context.Context, id uuid.UUID, userID uuid.UUID, provider string, providerUserID string) (db.OauthAccount, error) {
	return r.queries.CreateOAuthAccount(ctx, db.CreateOAuthAccountParams{
		ID:             id,
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	})
}

func (r *oauthAccountRepository) GetOAuthAccountByProviderAndUserID(ctx context.Context, provider string, providerUserID string) (db.OauthAccount, error) {
	account, err := r.queries.GetOAuthAccountByProviderAndUserID(ctx, db.GetOAuthAccountByProviderAndUserIDParams{
		Provider:       provider,
		ProviderUserID: providerUserID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.OauthAccount{}, ErrOAuthAccountNotFound
		}
		return db.OauthAccount{}, err
	}
	return account, nil
}

func (r *oauthAccountRepository) GetOAuthAccountByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) (db.OauthAccount, error) {
	account, err := r.queries.GetOAuthAccountByUserIDAndProvider(ctx, db.GetOAuthAccountByUserIDAndProviderParams{
		UserID:   userID,
		Provider: provider,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.OauthAccount{}, ErrOAuthAccountNotFound
		}
		return db.OauthAccount{}, err
	}
	return account, nil
}

func (r *oauthAccountRepository) GetOAuthAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]db.OauthAccount, error) {
	return r.queries.GetOAuthAccountsByUserID(ctx, userID)
}

func (r *oauthAccountRepository) DeleteOAuthAccount(ctx context.Context, id uuid.UUID) (db.OauthAccount, error) {
	return r.queries.DeleteOAuthAccount(ctx, id)
}

func (r *oauthAccountRepository) DeleteOAuthAccountByUserIDAndProvider(ctx context.Context, userID uuid.UUID, provider string) error {
	return r.queries.DeleteOAuthAccountByUserIDAndProvider(ctx, db.DeleteOAuthAccountByUserIDAndProviderParams{
		UserID:   userID,
		Provider: provider,
	})
}

// oauthAuthorizationCodeRepository implements OAuthAuthorizationCodeRepository.
type oauthAuthorizationCodeRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewOAuthAuthorizationCodeRepository creates a new OAuth authorization code repository.
func NewOAuthAuthorizationCodeRepository(db *sql.DB, queries *db.Queries) OAuthAuthorizationCodeRepository {
	return &oauthAuthorizationCodeRepository{db: db, queries: queries}
}

func (r *oauthAuthorizationCodeRepository) CreateAuthorizationCode(ctx context.Context, id uuid.UUID, codeHash string, userID uuid.NullUUID, nonce string, purpose string, expiresAt time.Time) (db.OauthAuthorizationCode, error) {
	return r.queries.CreateAuthorizationCode(ctx, db.CreateAuthorizationCodeParams{
		ID:        id,
		CodeHash:  codeHash,
		UserID:    userID,
		Nonce:     nonce,
		Purpose:   purpose,
		ExpiresAt: expiresAt,
	})
}

func (r *oauthAuthorizationCodeRepository) GetAuthorizationCodeByHash(ctx context.Context, codeHash string) (db.OauthAuthorizationCode, error) {
	code, err := r.queries.GetAuthorizationCodeByHash(ctx, codeHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.OauthAuthorizationCode{}, ErrOAuthCodeNotFound
		}
		return db.OauthAuthorizationCode{}, err
	}
	return code, nil
}

// GetAndLockAuthorizationCode acquires a row-level lock on the authorization code using
// SELECT FOR UPDATE within a transaction. This prevents concurrent code redemption.
// All error cases (not found, already used, expired) return ErrOAuthCodeNotFound
// to avoid leaking timing information about code state.
func (r *oauthAuthorizationCodeRepository) GetAndLockAuthorizationCode(ctx context.Context, codeHash string) (db.OauthAuthorizationCode, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return db.OauthAuthorizationCode{}, ErrOAuthCodeNotFound
	}
	defer func() { _ = tx.Rollback() }()

	code, err := r.lockAuthorizationCodeInTx(ctx, tx, codeHash)
	if err != nil {
		return db.OauthAuthorizationCode{}, ErrOAuthCodeNotFound
	}

	// Validate usage and expiry. All failures return the same error to prevent
	// timing oracles — the caller will only see ErrOAuthCodeNotFound.
	if code.UsedAt.Valid {
		return db.OauthAuthorizationCode{}, ErrOAuthCodeNotFound
	}
	if time.Now().After(code.ExpiresAt) {
		return db.OauthAuthorizationCode{}, ErrOAuthCodeNotFound
	}

	// Commit the transaction and return the valid code.
	if err := tx.Commit(); err != nil {
		return db.OauthAuthorizationCode{}, ErrOAuthCodeNotFound
	}
	return code, nil
}

// lockAuthorizationCodeInTx executes SELECT FOR UPDATE within the given transaction.
func (r *oauthAuthorizationCodeRepository) lockAuthorizationCodeInTx(ctx context.Context, tx *sql.Tx, codeHash string) (db.OauthAuthorizationCode, error) {
	const query = `
		SELECT id, code_hash, user_id, nonce, purpose, expires_at, used_at, created_at
		FROM oauth_authorization_codes
		WHERE code_hash = $1
		FOR UPDATE
	`
	var code db.OauthAuthorizationCode
	err := tx.QueryRowContext(ctx, query, codeHash).Scan(
		&code.ID,
		&code.CodeHash,
		&code.UserID,
		&code.Nonce,
		&code.Purpose,
		&code.ExpiresAt,
		&code.UsedAt,
		&code.CreatedAt,
	)
	if err != nil {
		return db.OauthAuthorizationCode{}, err
	}
	return code, nil
}

func (r *oauthAuthorizationCodeRepository) MarkAuthorizationCodeAsUsed(ctx context.Context, codeHash string) (db.OauthAuthorizationCode, error) {
	code, err := r.queries.MarkAuthorizationCodeAsUsed(ctx, codeHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.OauthAuthorizationCode{}, ErrOAuthCodeNotFound
		}
		return db.OauthAuthorizationCode{}, err
	}
	return code, nil
}

func (r *oauthAuthorizationCodeRepository) CleanupExpiredAuthorizationCodes(ctx context.Context) error {
	_, err := r.queries.DeleteExpiredAuthorizationCodes(ctx, time.Now())
	return err
}
